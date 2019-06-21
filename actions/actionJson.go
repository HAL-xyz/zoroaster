package actions

import (
	"encoding/json"
	"fmt"
	"zoroaster/triggers"
)

type ActionEventJson struct {
	ZTx     *trigger.ZTransaction
	Actions []string
}

type ActionEvent struct {
	ZTx     *trigger.ZTransaction
	Actions []Action
}

type Action struct {
	TriggerID  int
	UserID     int
	ActionType string
	Attribute  interface{}
}

type AttributeWebhookPost struct {
	URI string
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
	TriggerID  int    `json:"TriggerId"`
	UserID     int    `json:"UserId"`
	ActionType string `json:"ActionType"`
	Attributes struct {
		URI string `json:"URI"`
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
		TriggerID:  ajs.TriggerID,
		UserID:     ajs.UserID,
		ActionType: ajs.ActionType,
	}

	if ajs.ActionType == "webhook_post" {
		action.Attribute = AttributeWebhookPost{URI: ajs.Attributes.URI}
	} else {
		return nil, fmt.Errorf("invalid ActionType %s", ajs.ActionType)
	}
	return &action, nil
}
