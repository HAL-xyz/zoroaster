package actions

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewActionJson(t *testing.T) {
	var s = `{"UserId": 1, "TriggerId": 30, "ActionType": "webhook_post", "Attributes": {"URI": "https://webhook.site/202d0fac-4bfa-43f5-8ad0-c791cf051e5f"}}`
	_, err := NewActionJson([]byte(s))
	if err != nil {
		t.Error(err)
	}
}

func TestAction_UnmarshalJSON(t *testing.T) {
	var s = `{"UserId": 1, "TriggerId": 30, "ActionType": "webhook_post", "Attributes": {"URI": "https://webhook.site/202d0fac-4bfa-43f5-8ad0-c791cf051e5f"}}`
	a := Action{}

	err := json.Unmarshal([]byte(s), &a)
	if err != nil {
		t.Error(err)
	}

	_, ok := a.Attribute.(AttributeWebhookPost)
	assert.True(t, ok)
}
