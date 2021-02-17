package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/HAL-xyz/zoroaster/utils"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type IHttpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func ProcessActions(
	actionsString []string,
	match trigger.IMatch,
	iEmail sesiface.SESAPI,
	httpCli IHttpClient) []*trigger.Outcome {

	actions := getActionsFromString(actionsString)
	outcomes := make([]*trigger.Outcome, len(actions))

	var out = &trigger.Outcome{}
	for i, a := range actions {
		switch v := a.Attribute.(type) {
		case AttributeWebhookPost:
			out = handleWebHookPost(v, match, httpCli)
		case AttributeEmail:
			out = handleEmail(v, match, iEmail, a.TemplateVersion)
		case AttributeSlackBot:
			out = handleSlackBot(v, match, httpCli, a.TemplateVersion)
		case AttributeTelegramBot:
			out = handleTelegramBot(v, match, httpCli, a.TemplateVersion)
		case AttributeTweet:
			out = handleTweet(v, match, a.TemplateVersion)
		case AttributeDiscord:
			out = handleDiscord(v, match, httpCli, a.TemplateVersion)
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

func handleWebHookPost(awp AttributeWebhookPost, match trigger.IMatch, httpCli IHttpClient) *trigger.Outcome {

	postData, err := json.Marshal(match.ToPostPayload())
	if err != nil {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%v", match.ToPostPayload()), // bc the marshaling failed
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
		Success: resp.StatusCode == 200,
	}
}

type DiscordPayload struct {
	Content string `json:"content"`
}

func handleDiscord(discAttr AttributeDiscord, match trigger.IMatch, httpCli IHttpClient, templVersion string) *trigger.Outcome {
	payload := DiscordPayload{fillBodyTemplate(discAttr.Body, match, templVersion)}

	postData, err := json.Marshal(payload)
	if err != nil {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%s", payload),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}
	resp, err := httpCli.Post(discAttr.DiscordURI, "application/json", bytes.NewBuffer(postData))
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
		Success: resp.StatusCode == 204 || resp.StatusCode == 200, // Discord returns "No Content" instead of 200
	}
}

type SlackPayload struct {
	Text string `json:"text"`
}

func handleSlackBot(slackAttr AttributeSlackBot, match trigger.IMatch, httpCli IHttpClient, templVersion string) *trigger.Outcome {
	payload := SlackPayload{fillBodyTemplate(slackAttr.Body, match, templVersion)}

	postData, err := json.Marshal(payload)
	if err != nil {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%s", payload),
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
		Success: resp.StatusCode == 200,
	}
}

type TelegramPayload struct {
	ChatId       string `json:"chat_id"`
	Text         string `json:"text"`
	Format       string `json:"parse_mode"`
	LinksPreview bool   `json:"disable_web_page_preview"`
}

type TelegramResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

func handleTelegramBot(telegramAttr AttributeTelegramBot, match trigger.IMatch, httpCli IHttpClient, templVersion string) *trigger.Outcome {
	payload := TelegramPayload{
		Text:         fillBodyTemplate(telegramAttr.Body, match, templVersion),
		ChatId:       telegramAttr.ChatId,
		Format:       telegramAttr.Format,
		LinksPreview: telegramAttr.DisableLinksPreview,
	}

	// TODO this logic should be moved to the json unmarshaler
	chatIdRegex := regexp.MustCompile(`-\d+$|@.+|\d+$`)
	if !chatIdRegex.MatchString(payload.ChatId) {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%v", payload),
			Outcome: makeErrorResponse("Invalid chat ID"),
			Success: false,
		}
	}

	// TODO this logic should be moved to the json unmarshaler
	validFormats := []string{"Markdown", "MarkdownV2", "HTML"}
	if !utils.IsIn(payload.Format, validFormats) {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%v", payload),
			Outcome: makeErrorResponse("Invalid formatting directive"),
			Success: false,
		}
	}

	postData, err := json.Marshal(payload)
	if err != nil {
		return &trigger.Outcome{
			Payload: fmt.Sprintf("%v", payload),
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

	apiOutcome := TelegramResponse{}
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &apiOutcome)
	if err != nil {
		return &trigger.Outcome{
			Payload: string(postData),
			Outcome: makeErrorResponse(err.Error()),
			Success: false,
		}
	}

	persistentResp := WebhookResponse{resp.StatusCode, resp.Status}
	if !apiOutcome.Ok {
		persistentResp.Response = apiOutcome.Description
	}
	persistentJsonResp, _ := json.Marshal(persistentResp)

	return &trigger.Outcome{
		Payload: string(postData),
		Outcome: string(persistentJsonResp),
		Success: resp.StatusCode == 200,
	}
}

type TwitterPayload struct {
	Status string
}

func handleTweet(tweetAttr AttributeTweet, match trigger.IMatch, templVersion string) *trigger.Outcome {
	payload := TwitterPayload{
		Status: fillBodyTemplate(tweetAttr.Status, match, templVersion),
	}

	postData, _ := json.Marshal(payload)

	authconfig := oauth1.NewConfig(config.Zconf.TwitterConsumerKey, config.Zconf.TwitterConsumerSecret)
	token := oauth1.NewToken(tweetAttr.Token, tweetAttr.Secret)
	httpClient := authconfig.Client(oauth1.NoContext, token)
	twitterClient := twitter.NewClient(httpClient)

	_, resp, err := twitterClient.Statuses.Update(payload.Status, nil)

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

func handleEmail(email AttributeEmail, match trigger.IMatch, iemail sesiface.SESAPI, templVersion string) *trigger.Outcome {

	email.Body = fillBodyTemplate(email.Body, match, templVersion)
	email.Subject = fillBodyTemplate(email.Subject, match, templVersion)
	allRecipients := getAllRecipients(email.To, match, templVersion)

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
func getAllRecipients(emailTo []string, match trigger.IMatch, templVersion string) []string {
	extraRecipients := make([]string, 0)
	extraRecipients = append(extraRecipients, emailTo...)

	for _, r := range emailTo {
		templatedString := fillBodyTemplate(r, match, templVersion)
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
