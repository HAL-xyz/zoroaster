package aws

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/onrik/ethrpc"
	"log"
	"os"
	"time"
	"zoroaster/triggers"
)

var db *sql.DB

func LogMatch(tg *trigger.Trigger, tx *ethrpc.Transaction, logTable string) {
	q := fmt.Sprintf(`INSERT INTO "%s" ("date", "trigger_id", "block_no", "tx_hash") VALUES($1, $2, $3, $4)`, logTable)
	_, err := db.Exec(q, time.Now(), tg.TriggerId, *tx.BlockNumber, tx.Hash)
	if err != nil {
		log.Printf("WARN: Cannot write trigger log match: %s", err)
	}
}

func LoadTriggersFromDB(table string) ([]*trigger.Trigger, error) {
	sqlSt := fmt.Sprintf("SELECT trigger_data FROM %s", table)
	rows, err := db.Query(sqlSt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	triggers := make([]*trigger.Trigger, 0)
	for rows.Next() {
		var tg string
		err = rows.Scan(&tg)
		if err != nil {
			panic(err)
		}
		trig, err := trigger.NewTriggerFromJson(tg)
		if err != nil {
			log.Println(err)
		} else {
			triggers = append(triggers, trig)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return triggers, nil
}

func InitDB() {
	const (
		dbEndpoint = "triggersdb-test.cgylkhaks4ty.eu-central-1.rds.amazonaws.com"
		dbUser     = "triggersdb_test_admin"
		dbName     = "triggers"
		pwdEnv     = "ZORO_PASS"
	)
	pwd := os.Getenv(pwdEnv)
	if pwd == "" {
		log.Fatal("No db password set in local env ", pwdEnv)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbEndpoint, 5432, dbUser, pwd, dbName)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot connect to the DB -> ", err)
	}
}
