package matcher

import (
	"github.com/HAL-xyz/zoroaster/action"
	"github.com/HAL-xyz/zoroaster/db"
	"github.com/HAL-xyz/zoroaster/trigger"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type IHttpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func ProcessMatch(match trigger.IMatch, idb db.IDB, iEmail sesiface.SESAPI, httpCli IHttpClient) []*trigger.Outcome {

	acts, err := idb.GetActions(match.GetTriggerUUID(), match.GetUserUUID())
	if err != nil {
		log.Fatalf("cannot get actions from db: %v", err)
	}
	log.Debugf("\tMatched %d actions", len(acts))

	outcomes := action.ProcessActions(acts, match, iEmail, httpCli)
	for _, out := range outcomes {
		err = idb.LogOutcome(out, match.GetMatchUUID())
		if err != nil {
			log.Error(err)
		}
		log.Debug("\tLogged outcome for match id ", match.GetMatchUUID())
	}
	return outcomes
}
