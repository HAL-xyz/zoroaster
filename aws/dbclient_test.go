package aws

import (
	"log"
	"testing"
	"zoroaster/config"
	"zoroaster/trigger"
)

var psqlClient = PostgresClient{}
var zconf = config.Load("../config")

func init() {
	if zconf.Stage != "DEV" {
		log.Fatal("$STAGE must be DEV to run db tests")
	}
	psqlClient.InitDB(zconf)
}

func TestPostgresClient_LogCnMatch(t *testing.T) {
	defer psqlClient.Close()

	m := trigger.CnMatch{1, 8888, 10, "xxx xxx xxx"}

	psqlClient.LogCnMatch(zconf.TriggersDB.TableCnMatches, m)
}
