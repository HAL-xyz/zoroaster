package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"zoroaster/triggers"
)

func HandleEvent(evJson ActionEventJson) []*trigger.Outcome {
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

	outcomes := make([]*trigger.Outcome, len(event.Actions))
	for i, a := range event.Actions {
		switch v := a.Attribute.(type) {
		case AttributeWebhookPost:
			out := handleWebHookPost(v, event.ZTx)
			outcomes[i] = out
		default:
			out := &trigger.Outcome{fmt.Sprintf("unsupported ActionType: %s", a.ActionType), ""}
			outcomes[i] = out
		}
	}
	return outcomes
}

func handleWebHookPost(awp AttributeWebhookPost, ztx *trigger.ZTransaction) *trigger.Outcome {
	dataBytes, err := json.Marshal(*ztx)
	if err != nil {
		return &trigger.Outcome{err.Error(), ""}
	}
	resp, err := http.Post(awp.URI, "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		return &trigger.Outcome{err.Error(), string(dataBytes)}
	}
	return &trigger.Outcome{resp.Status, string(dataBytes)}
}
