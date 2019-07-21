package matcher

import (
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	log "github.com/sirupsen/logrus"
	"zoroaster/action"
	"zoroaster/aws"
	"zoroaster/config"
	"zoroaster/trigger"
)

func ProcessMatch(match interface{}, idb aws.IDB, zconf *config.ZConfiguration, iemail sesiface.SESAPI) {

	switch m := match.(type) {
	case *trigger.TxMatch:
		log.Debug("Got a Tx Match")
		acts, err := idb.GetActions(zconf.TriggersDB.TableActions, m.Tg.TriggerId, m.Tg.UserId)
		if err != nil {
			log.Warnf("cannot get actions from db: %v", err)
		}
		log.Debugf("\tMatched %d actions", len(acts))

		outcomes := action.ProcessActions(acts, m.ZTx, iemail)
		for _, out := range outcomes {
			idb.LogOutcome(zconf.TriggersDB.TableOutcomes, out, m.MatchId)
			log.Debug("\tLogged outcome for match id ", m.MatchId)
		}
	case *trigger.CnMatch:
		log.Debug("Got a Contract Match")
		acts, err := idb.GetActions(zconf.TriggersDB.TableActions, m.TgId, m.TgUserId)
		if err != nil {
			log.Warnf("cannot get actions from db: %v", err)
		}
		log.Debugf("\tMatched %d actions", len(acts))

		outcomes := action.ProcessActions(acts, m.Value, iemail)
		for _, out := range outcomes {
			idb.LogOutcome(zconf.TriggersDB.TableOutcomes, out, m.MatchId)
			log.Debug("\tLogged outcome for match id ", m.MatchId)
		}
	default:
		log.Fatalf("invalid match type %T", m)
	}
}
