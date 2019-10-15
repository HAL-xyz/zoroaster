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
		return &trigger.Outcome{fmt.Sprintf("%s", emailPayload), err.Error()}
	}

	result, err := sendEmail(iemail, allRecipients, email.Subject, body)
	if err != nil {
		return &trigger.Outcome{string(emailPayloadString), err.Error()}
	}
	return &trigger.Outcome{string(emailPayloadString), result.String()}
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
