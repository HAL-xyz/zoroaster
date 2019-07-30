package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	log "github.com/sirupsen/logrus"
	"zoroaster/aws"
	"zoroaster/trigger"
)

func ProcessActions(
	actionsString []string,
	payload interface{},
	iEmail sesiface.SESAPI,
	httpCli aws.IHttpClient) []*trigger.Outcome {

	actions := getActionsFromString(actionsString)
	outcomes := make([]*trigger.Outcome, len(actions))

	var out = &trigger.Outcome{}
	for i, a := range actions {
		switch v := a.Attribute.(type) {
		case AttributeWebhookPost:
			out = handleWebHookPost(v, payload, httpCli)
		case AttributeEmail:
			out = handleEmail(v, payload, iEmail)
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

func handleWebHookPost(awp AttributeWebhookPost, payload interface{}, httpCli aws.IHttpClient) *trigger.Outcome {
	m, ok := payload.(*trigger.CnMatch)
	if ok {
		payload = toCnPostData(m)
	}

	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return &trigger.Outcome{err.Error(), ""}
	}
	resp, err := httpCli.Post(awp.URI, "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		return &trigger.Outcome{err.Error(), string(dataBytes)}
	}
	return &trigger.Outcome{resp.Status, string(dataBytes)}
}

func handleEmail(email AttributeEmail, paylaod interface{}, iemail sesiface.SESAPI) *trigger.Outcome {
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

// In the web hook POST we don't want to expose TgId and TgUserId
// so we send this instead of a full CnMatch
type ContractPostData struct {
	MatchId        int
	BlockNo        int
	ReturnValue    string
	BlockTimestamp int
}

func toCnPostData(m *trigger.CnMatch) *ContractPostData {
	return &ContractPostData{
		MatchId:        m.MatchId,
		BlockNo:        m.BlockNo,
		ReturnValue:    m.Value,
		BlockTimestamp: m.BlockTimestamp,
	}
}
