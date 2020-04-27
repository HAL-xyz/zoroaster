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
		case AttributeSlackBot:
			out = handleSlackBot(v, match, httpCli)
		default:
			out = &trigger.Outcome{
				Payload: "",
				Outcome: fmt.Sprintf("unsupported ActionType: %s", a.ActionType),
				Success: false,
			}
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

type ErrorMsg struct {
	Error string `json:"error"`
}

func makeErrorResponse(e string) string {
	errMsg := ErrorMsg{e}
	errJsn, _ := json.Marshal(errMsg)
	return string(errJsn)
}

type WebhookResponse struct {
	HttpCode int
	Response string
}

func handleWebHookPost(awp AttributeWebhookPost, match trigger.IMatch, httpCli aws.IHttpClient) *trigger.Outcome {

	postData, err := json.Marshal(match.ToPostPayload())
	if err != nil {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%v", match.ToPostPayload()),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}
	resp, err := httpCli.Post(awp.URI, "application/json", bytes.NewBuffer(postData))
	if err != nil {
		return &trigger.Outcome{
			Payload: string(postData),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}
	defer resp.Body.Close()

	responseCode := WebhookResponse{resp.StatusCode, resp.Status}
	jsonRespCode, _ := json.Marshal(responseCode)
	return &trigger.Outcome{
		Payload: string(postData),
		Outcome: string(jsonRespCode),
		Success: true,
	}
}

type SlackPayload struct {
	Text string `json:"text"`
}

func handleSlackBot(slackAttr AttributeSlackBot, match trigger.IMatch, httpCli aws.IHttpClient) *trigger.Outcome {
	txt := SlackPayload{fillBodyTemplate(slackAttr.Body, match)}

	postData, err := json.Marshal(txt)
	if err != nil {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%s", slackAttr.Body),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}
	resp, err := httpCli.Post(slackAttr.URI, "application/json", bytes.NewBuffer(postData))
	if err != nil {
		return &trigger.Outcome{
			Payload: string(postData),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}
	defer resp.Body.Close()

	responseCode := WebhookResponse{resp.StatusCode, resp.Status}
	jsonRespCode, _ := json.Marshal(responseCode)
	return &trigger.Outcome{
		Payload: string(postData),
		Outcome: string(jsonRespCode),
		Success: true,
	}
}

type TelegramPayload struct {
	ChatId string `json:"chat_id"`
	Text   string `json:"text"`
	Format string `json:"parse_mode"`
}

func handleTelegramBot(telegramAttr AttributeTelegramBot, match trigger.IMatch, httpCli aws.IHttpClient) *trigger.Outcome {
	payload := TelegramPayload{
		Text:   fillBodyTemplate(telegramAttr.Body, match),
		ChatId: telegramAttr.ChatId,
		Format: telegramAttr.Format,
	}

	if !(strings.HasPrefix(payload.ChatId, "-") || (strings.HasPrefix(payload.ChatId, "@"))) {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%s", payload),
			Outcome: makeErrorResponse("Invalid chat ID"),
			Success: false,
		}
	}

	validFormats := []string{"Markdown", "MarkdownV2", "HTML"}
	if !utils.IsIn(payload.Format, validFormats) {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%s", payload),
			Outcome: makeErrorResponse("Invalid formatting directive"),
			Success: false,
		}
	}

	postData, err := json.Marshal(payload)
	if err != nil {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%s", payload),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}

	URI := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramAttr.Token)

	resp, err := httpCli.Post(URI, "application/json", bytes.NewBuffer(postData))
	if err != nil {
		return &trigger.Outcome{
			Payload: string(postData),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}
	defer resp.Body.Close()

	responseCode := WebhookResponse{resp.StatusCode, resp.Status}
	jsonRespCode, _ := json.Marshal(responseCode)
	return &trigger.Outcome{
		Payload: string(postData),
		Outcome: string(jsonRespCode),
		Success: resp.StatusCode == 200,
	}
}

type EmailPayload struct {
	Recipients []string
	Body       string
	Subject    string
}

func handleEmail(email AttributeEmail, match trigger.IMatch, iemail sesiface.SESAPI) *trigger.Outcome {

	email.Body = fillBodyTemplate(email.Body, match)
	email.Subject = fillBodyTemplate(email.Subject, match)
	allRecipients := getAllRecipients(email.To, match)

	emailPayload := EmailPayload{
		Recipients: allRecipients,
		Body:       email.Body,
		Subject:    email.Subject,
	}
	emailPayloadJson, err := json.Marshal(emailPayload)
	if err != nil {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%s", emailPayload),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}
	result, err := sendEmail(iemail, allRecipients, email.Subject, email.Body)
	if err != nil {
		return &trigger.Outcome{
			Payload: string(emailPayloadJson),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}
	outcomeJsn, _ := json.Marshal(result)
	return &trigger.Outcome{
		Payload: string(emailPayloadJson),
		Outcome: string(outcomeJsn),
		Success: true,
	}
}

// get extra recipients from the TO field
func getAllRecipients(emailTo []string, match trigger.IMatch) []string {
	extraRecipients := make([]string, 0)
	extraRecipients = append(extraRecipients, emailTo...)

	for _, r := range emailTo {
		templatedString := fillBodyTemplate(r, match)
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
