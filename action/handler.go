package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ses"
	log "github.com/sirupsen/logrus"
	"net/http"
	"zoroaster/trigger"
)

func HandleEvent(evJson ActionEventJson, sess *ses.SES) []*trigger.Outcome {
	event := ActionEvent{}
	event.ZTx = evJson.ZTx

	for _, a := range evJson.Actions {
		act := Action{}
		err := json.Unmarshal([]byte(a), &act)
		if err != nil {
			log.Debug(err)
			continue
		}
		event.Actions = append(event.Actions, act)
	}

	outcomes := make([]*trigger.Outcome, len(event.Actions))
	var out = &trigger.Outcome{}

	for i, a := range event.Actions {
		switch v := a.Attribute.(type) {
		case AttributeWebhookPost:
			out = handleWebHookPost(v, event.ZTx)
		case AttributeEmail:
			out = handleEmail(sess, v, event.ZTx)
		default:
			out = &trigger.Outcome{fmt.Sprintf("unsupported ActionType: %s", a.ActionType), ""}
		}
		outcomes[i] = out
	}
	return outcomes
}

func handleWebHookPost(awp AttributeWebhookPost, ztx *trigger.ZTransaction) *trigger.Outcome {
	dataBytes, err := json.Marshal(*ztx)
	if err != nil {
		return &trigger.Outcome{err.Error(), ""}
	}
	resp, err := http.Post(awp.URI, "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		return &trigger.Outcome{err.Error(), string(dataBytes)}
	}
	return &trigger.Outcome{resp.Status, string(dataBytes)}
}

func handleEmail(sess *ses.SES, email AttributeEmail, ztx *trigger.ZTransaction) *trigger.Outcome {
	body := FillEmailTemplate(email.Body, ztx)

	result, err := sendEmail(sess, email.To, email.Subject, body)
	if err != nil {
		return &trigger.Outcome{err.Error(), ""}
	}
	return &trigger.Outcome{result.String(), email.Body}
}
