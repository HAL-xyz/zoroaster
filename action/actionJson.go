package action

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Action struct {
	TriggerID       int
	UserID          int
	ActionType      string
	Attribute       interface{}
	TemplateVersion string
}

type AttributeWebhookPost struct {
	URI string
}

type AttributeDiscord struct {
	DiscordURI string
	Body       string
}

type AttributeEmail struct {
	From    string
	To      []string
	Subject string
	Body    string
}

type AttributeSlackBot struct {
	URI  string
	Body string
}

type AttributeTelegramBot struct {
	Body                string
	Token               string
	ChatId              string
	Format              string
	DisableLinksPreview bool
}

type AttributeTweet struct {
	Status string
	Token  string
	Secret string
}

// Implements the json.Unmarshaler interface
func (a *Action) UnmarshalJSON(data []byte) error {
	proxy, err := NewActionJson(data)
	if err != nil {
		return err
	}
	ret, err := proxy.ToAction()
	if err != nil {
		return err
	}
	*a = *ret

	return err
}

// proxy struct
type ActionJson struct {
	TriggerID       int    `json:"TriggerUUID"`
	UserID          int    `json:"UserUUID"`
	ActionType      string `json:"ActionType"`
	TemplateVersion string `json:"TemplateVersion"`
	Attributes      struct {
		URI                         string   `json:"URI"`
		To                          []string `json:"To"`
		Subject                     string   `json:"Subject"`
		Body                        string   `json:"Body"`
		ChatId                      string   `json:"ChatId"`
		Token                       string   `json:"Token"`
		Secret                      string   `json:"Secret"`
		Status                      string   `json:"Status"`
		Format                      string   `json:"Format"`
		DiscordURI                  string   `json:"DiscordURI"`
		DisableTelegramLinksPreview *bool    `json:"DisableTelegramLinksPreview",omitempty`
	} `json:"Attributes"`
}

// creates a new ActionJson from JSON
func NewActionJson(input []byte) (*ActionJson, error) {
	aj := ActionJson{}
	err := json.Unmarshal([]byte(string(input)), &aj)
	if err != nil {
		return nil, err
	}
	return &aj, nil
}

// converts an ActionJson to an Action
func (ajs *ActionJson) ToAction() (*Action, error) {
	action := Action{
		TriggerID:       ajs.TriggerID,
		UserID:          ajs.UserID,
		ActionType:      ajs.ActionType,
		TemplateVersion: ajs.TemplateVersion,
	}

	switch strings.ToLower(ajs.ActionType) {
	case "webhook_post":
		action.Attribute = AttributeWebhookPost{URI: ajs.Attributes.URI}
	case "email":
		action.Attribute = AttributeEmail{
			To:      ajs.Attributes.To,
			Subject: ajs.Attributes.Subject,
			Body:    ajs.Attributes.Body,
		}
	case "slack":
		action.Attribute = AttributeSlackBot{
			URI:  ajs.Attributes.URI,
			Body: ajs.Attributes.Body,
		}
	case "telegram":
		disableLinksPreview := true
		if ajs.Attributes.DisableTelegramLinksPreview != nil {
			disableLinksPreview = *(ajs.Attributes.DisableTelegramLinksPreview)
		}
		action.Attribute = AttributeTelegramBot{
			Token:               ajs.Attributes.Token,
			Body:                ajs.Attributes.Body,
			ChatId:              ajs.Attributes.ChatId,
			Format:              ajs.Attributes.Format,
			DisableLinksPreview: disableLinksPreview,
		}
	case "twitter":
		action.Attribute = AttributeTweet{
			Token:  ajs.Attributes.Token,
			Secret: ajs.Attributes.Secret,
			Status: ajs.Attributes.Status,
		}
	case "discord":
		action.Attribute = AttributeDiscord{
			DiscordURI: ajs.Attributes.DiscordURI,
			Body:       ajs.Attributes.Body,
		}

	default:
		return nil, fmt.Errorf("invalid ActionType %s", ajs.ActionType)
	}
	return &action, nil
}
