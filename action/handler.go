package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"zoroaster/aws"
	"zoroaster/trigger"
	"zoroaster/utils"
)

func ProcessActions(
	actionsString []string,
	payload trigger.IMatch,
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

func handleWebHookPost(awp AttributeWebhookPost, match trigger.IMatch, httpCli aws.IHttpClient) *trigger.Outcome {
	var payload interface{}

	m, ok := match.(*trigger.CnMatch)
	if ok {
		payload = toCnPostData(m)
	} else {
		payload = match
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

func handleEmail(email AttributeEmail, payload trigger.IMatch, iemail sesiface.SESAPI) *trigger.Outcome {
	body := fillEmailTemplate(email.Body, payload)

	// get extra recipients from the TO field
	extraRecipients := make([]string, 0)
	for _, r := range email.To {
		newAddress := fillEmailTemplate(r, payload)
		// we need to figure out what type the return value is
		returnType := "single"
		if strings.HasPrefix(newAddress, "[") {
			returnType = "array"
		}
		if strings.Contains(newAddress, " ") && !strings.HasPrefix(newAddress, "[") {
			returnType = "multiple"
		}
		switch returnType {
		case "array":
			emails := strings.Split(newAddress, ",")
			for _, em := range emails {
				em = utils.RemoveCharacters(em, "[]")
				if !utils.IsIn(em, email.To) {
					extraRecipients = append(extraRecipients, em)
				}
			}
		case "multiple":
			emails := strings.Split(newAddress, " ")
			for _, em := range emails {
				if !utils.IsIn(em, email.To) {
					extraRecipients = append(extraRecipients, em)
				}
			}
		case "single":
			if !utils.IsIn(newAddress, email.To) {
				extraRecipients = append(extraRecipients, newAddress)
			}
		}
	}
	allRecipients := append(email.To, extraRecipients...)
	allRecipients = validateEmails(allRecipients)

	emailPayload := trigger.EmailPayload{
		Recipients: allRecipients,
		Body:       body,
	}
	emailPayloadString, err := json.Marshal(emailPayload)
	if err != nil {
		return &trigger.Outcome{err.Error(), ""}
	}

	result, err := sendEmail(iemail, allRecipients, email.Subject, body)
	if err != nil {
		return &trigger.Outcome{err.Error(), ""}
	}
	return &trigger.Outcome{result.String(), string(emailPayloadString)}
}

func validateEmails(emails []string) []string {
	rg := regexp.MustCompile(`\w*@\w.\w`)
	validEmails := make([]string, 0)
	for _, email := range emails {
		if rg.MatchString(email) {
			validEmails = append(validEmails, email)
		}
	}
	return validEmails
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
		ReturnValue:    m.MatchedValues,
		BlockTimestamp: m.BlockTimestamp,
	}
}
