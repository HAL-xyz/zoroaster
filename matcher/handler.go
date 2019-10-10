package matcher

import (
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	log "github.com/sirupsen/logrus"
	"zoroaster/action"
	"zoroaster/aws"
	"zoroaster/trigger"
)

func ProcessMatch(match trigger.IMatch, idb aws.IDB, iEmail sesiface.SESAPI, httpCli aws.IHttpClient) []*trigger.Outcome {

	var userId, triggerId, matchId int

	switch m := match.(type) {
	case trigger.TxMatch:
		log.Debug("Got a Tx Match")
		userId, triggerId, matchId = m.Tg.UserId, m.Tg.TriggerId, m.MatchId
	case trigger.CnMatch:
		log.Debug("Got a Contract Match")
		userId, triggerId, matchId = m.Trigger.UserId, m.Trigger.TriggerId, m.MatchId
	default:
		log.Errorf("invalid match type %T", m)
		return nil
	}
	acts, err := idb.GetActions(triggerId, userId)
	if err != nil {
		log.Warnf("cannot get actions from db: %v", err)
	}
	log.Debugf("\tMatched %d actions", len(acts))

	outcomes := action.ProcessActions(acts, match, iEmail, httpCli)
	for _, out := range outcomes {
		idb.LogOutcome(out, matchId)
		log.Debug("\tLogged outcome for match id ", matchId)
	}
	return outcomes
}
