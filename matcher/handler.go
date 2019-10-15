package matcher

import (
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	log "github.com/sirupsen/logrus"
	"zoroaster/action"
	"zoroaster/aws"
	"zoroaster/trigger"
)

func ProcessMatch(match trigger.IMatch, idb aws.IDB, iEmail sesiface.SESAPI, httpCli aws.IHttpClient) []*trigger.Outcome {

	var userUUID, triggerUUID, matchUUID string

	switch m := match.(type) {
	case trigger.TxMatch:
		log.Debug("Got a Tx Match")
		userUUID, triggerUUID, matchUUID = m.Tg.UserUUID, m.Tg.TriggerUUID, m.MatchUUID
	case trigger.CnMatch:
		log.Debug("Got a Contract Match")
		userUUID, triggerUUID, matchUUID = m.Trigger.UserUUID, m.Trigger.TriggerUUID, m.MatchUUID
	default:
		log.Errorf("invalid match type %T", m)
		return nil
	}
	acts, err := idb.GetActions(triggerUUID, userUUID)
	if err != nil {
		log.Warnf("cannot get actions from db: %v", err)
	}
	log.Debugf("\tMatched %d actions", len(acts))

	outcomes := action.ProcessActions(acts, match, iEmail, httpCli)
	for _, out := range outcomes {
		idb.LogOutcome(out, matchUUID)
		log.Debug("\tLogged outcome for match id ", matchUUID)
	}
	return outcomes
}
