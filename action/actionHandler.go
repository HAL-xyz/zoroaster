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
	match trigger.IMatch,
	iEmail sesiface.SESAPI,
	httpCli aws.IHttpClient) []*trigger.Outcome {

	actions := getActionsFromString(actionsString)
	outcomes := make([]*trigger.Outcome, len(actions))

	var out = &trigger.Outcome{}
	for i, a := range actions {
		switch v := a.Attribute.(type) {
		case AttributeWebhookPost:
			out = handleWebHookPost(v, match, httpCli)
		case AttributeEmail:
			out = handleEmail(v, match, iEmail)
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

	postData, err := json.Marshal(match.ToPostPayload())
	if err != nil {
		return &trigger.Outcome{fmt.Sprintf("%v", match.ToPostPayload()), err.Error()}
	}
	resp, err := httpCli.Post(awp.URI, "application/json", bytes.NewBuffer(postData))
	if err != nil {
		return &trigger.Outcome{string(postData), err.Error()}
	}

	responseCode := trigger.WebhookResponse{resp.StatusCode}
	jsonRespCode, _ := json.Marshal(responseCode)
	return &trigger.Outcome{string(postData), string(jsonRespCode)}
}

func handleEmail(email AttributeEmail, match trigger.IMatch, iemail sesiface.SESAPI) *trigger.Outcome {

	body := fillEmailTemplate(email.Body, match)
	allRecipients := getAllRecipients(email.To, match)

	emailPayload := trigger.EmailPayload{
		Recipients: allRecipients,
		Body:       body,
	}
	emailPayloadJson, err := json.Marshal(emailPayload)
	if err != nil {
		errJson, _ := json.Marshal(err)
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%s", emailPayload),
			Outcome: string(errJson),
		}
	}
	result, err := sendEmail(iemail, allRecipients, email.Subject, body)
	if err != nil {
		errJson, _ := json.Marshal(err)
		return &trigger.Outcome{
			Payload: string(emailPayloadJson),
			Outcome: string(errJson),
		}
	}
	outcomeJsn, _ := json.Marshal(result)
	return &trigger.Outcome{
		Payload: string(emailPayloadJson),
		Outcome: string(outcomeJsn),
	}
}

// get extra recipients from the TO field
func getAllRecipients(emailTo []string, match trigger.IMatch) []string {
	extraRecipients := make([]string, 0)
	extraRecipients = append(extraRecipients, emailTo...)

	for _, r := range emailTo {
		templatedString := fillEmailTemplate(r, match)
		cleanString := utils.RemoveCharacters(templatedString, "[]")
		for _, email := range strings.Split(cleanString, " ") {
			if !utils.IsIn(email, extraRecipients) {
				extraRecipients = append(extraRecipients, email)
			}
		}
	}
	return validateEmails(extraRecipients)
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
