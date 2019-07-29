package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	log "github.com/sirupsen/logrus"
	"net/http"
	"zoroaster/trigger"
)

func ProcessActions(actionsString []string, payload interface{}, iemail sesiface.SESAPI) []*trigger.Outcome {
	actions := getActionsFromString(actionsString)
	outcomes := make([]*trigger.Outcome, len(actions))

	var out = &trigger.Outcome{}
	for i, a := range actions {
		switch v := a.Attribute.(type) {
		case AttributeWebhookPost:
			out = handleWebHookPost(v, payload)
		case AttributeEmail:
			out = handleEmail(iemail, v, payload)
		default:
			out = &trigger.Outcome{fmt.Sprintf("unsupported ActionType: %s", a.ActionType), ""}
		}
		outcomes[i] = out
	}
	return outcomes
}

func getActionsFromString(actionsString []string) []*Action {
	actions := make([]*Action, len(actionsString))
	for i, a := range actionsString {
		act := Action{}
		err := json.Unmarshal([]byte(a), &act)
		if err != nil {
			log.Debug(err)
			continue
		}
		actions[i] = &act
	}
	return actions
}

func handleWebHookPost(awp AttributeWebhookPost, payload interface{}) *trigger.Outcome {
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return &trigger.Outcome{err.Error(), ""}
	}
	resp, err := http.Post(awp.URI, "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		return &trigger.Outcome{err.Error(), string(dataBytes)}
	}
	return &trigger.Outcome{resp.Status, string(dataBytes)}
}

func handleEmail(iemail sesiface.SESAPI, email AttributeEmail, paylaod interface{}) *trigger.Outcome {
	var body string
	ztx, ok := paylaod.(*trigger.ZTransaction)
	if ok {
		body = fillEmailTemplate(email.Body, ztx)
	} else {
		body = email.Body
	}

	result, err := sendEmail(iemail, email.To, email.Subject, body)
	if err != nil {
		return &trigger.Outcome{err.Error(), ""}
	}
	return &trigger.Outcome{result.String(), email.Body}
}
