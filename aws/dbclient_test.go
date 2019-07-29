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

func TestPostgresClient_All(t *testing.T) {
	// TODO figure out how Go does teardown so I can split these tests;
	// for now I can't be bothered and I'll fit everything in one test,
	// closing the connection only once, at the end.
	defer psqlClient.Close()

	m := trigger.CnMatch{1, 8888, 10, 0, "xxx xxx xxx"}
	psqlClient.LogCnMatch(m)

	psqlClient.UpdateMatchingTriggers([]int{21, 31})

	psqlClient.UpdateNonMatchingTriggers([]int{21, 31})

	o1 := trigger.Outcome{"TX outcome", "TX payload"}
	o2 := trigger.Outcome{"CN outcome", "CN payload"}
	psqlClient.LogOutcome(&o1, 1, "wat")
	psqlClient.LogOutcome(&o2, 1, "wac")

}
