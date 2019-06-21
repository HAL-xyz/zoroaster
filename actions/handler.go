package actions

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"zoroaster/triggers"
)

func HandleEvent(evJson ActionEventJson) {
	event := ActionEvent{}
	event.ZTx = evJson.ZTx

	for _, a := range evJson.Actions {
		act := Action{}
		err := json.Unmarshal([]byte(a), &act)
		if err != nil {
			log.Debug(err)
			continue
		}
		event.Actions = append(event.Actions, act)
	}

	for _, a := range event.Actions {
		switch v := a.Attribute.(type) {
		case AttributeWebhookPost:
			handleWebHookPost(v, event.ZTx)
		default:
			log.Warn("unsupported ActionType:", a.ActionType)
			continue
		}
	}
}

func handleWebHookPost(awp AttributeWebhookPost, ztx *trigger.ZTransaction) {
	dataBytes, err := json.Marshal(*ztx)
	if err != nil {
		log.Warn(err)
		return
	}
	resp, err := http.Post(awp.URI, "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		log.Warn(err)
	} else {
		log.Debugf("\tAction sent - %s", resp.Status)
	}
}
