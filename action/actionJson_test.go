package action

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetProxyActionFromJson(t *testing.T) {
	var s = `{
   "UserUUID":1,
   "TriggerUUID":30,
   "ActionType":"webhook_post",
   "Attributes":{
      "URI":"https://webhook.site/202d0fac-4bfa-43f5-8ad0-c791cf051e5f"
   }
}`
	_, err := NewActionJson([]byte(s))
	assert.NoError(t, err)
}

func TestGetWebhookActionFromJson(t *testing.T) {
	var s = `{
   "UserUUID":1,
   "TriggerUUID":30,
   "ActionType":"webhook_post",
   "Attributes":{
      "URI":"https://webhook.site/202d0fac-4bfa-43f5-8ad0-c791cf051e5f"
   }
}`
	a := Action{}

	err := json.Unmarshal([]byte(s), &a)
	assert.NoError(t, err)

	_, ok := a.Attribute.(AttributeWebhookPost)
	assert.True(t, ok)
}

func TestGetEmailActionFromJson(t *testing.T) {
	var s = `
	{  
	   "UserUUID":1,
	   "TriggerUUID":30,
	   "ActionType":"email",
	   "Attributes":{  
		  "To":[  
			 "manlio.poltronieri@gmail.com",
			 "marco@atomic.eu.com"
		  ],
		  "Subject":"YO from Zoroaster",
		  "Body":"yo yo yo and a bottle of rum"
	   }
	}`
	a := Action{}

	err := json.Unmarshal([]byte(s), &a)
	assert.NoError(t, err)

	_, ok := a.Attribute.(AttributeEmail)
	assert.True(t, ok)
}

func TestGetSlackActionFromJson(t *testing.T) {
	var s = `
	{  
	   "UserUUID":1,
	   "TriggerUUID":30,
	   "ActionType":"slack",
	   "Attributes":{  
		  "URI":"https://hooks.slack.com/services/blahblah",
		  "Body":"just nod if you can here me"
	   }
	}`
	a := Action{}

	err := json.Unmarshal([]byte(s), &a)
	assert.NoError(t, err)

	_, ok := a.Attribute.(AttributeSlackBot)
	assert.True(t, ok)
}

func TestGetTelegramActionFromJson(t *testing.T) {
	var s = `
	{  
	   "UserUUID":1,
	   "TriggerUUID":30,
	   "ActionType":"telegram",
	   "Attributes":{  
		  "Token":"2932842309482309482394823",
		  "Body":"hey jude",
		  "ChatId": "408369342",
		  "Format": "MarkdownV2"
	   }
	}`
	a := Action{}

	err := json.Unmarshal([]byte(s), &a)
	assert.NoError(t, err)

	_, ok := a.Attribute.(AttributeTelegramBot)
	assert.True(t, ok)
}

func TestGetTwitterActionFromJson(t *testing.T) {
	var s = `
	{  
	   "UserUUID":1,
	   "TriggerUUID":30,
	   "ActionType":"twitter",
	   "Attributes":{  
		  "Token":"2329323098204983204983",
		  "Secret":"sssssssssssht",
		  "Body":"hey jude "
	   }
	}`
	a := Action{}

	err := json.Unmarshal([]byte(s), &a)
	assert.NoError(t, err)

	_, ok := a.Attribute.(AttributeTweet)
	assert.True(t, ok)
}
