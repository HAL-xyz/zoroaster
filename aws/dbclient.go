package aws

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"zoroaster/triggers"
)

var db *sql.DB

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
	log.Printf("=> Loaded %d triggers from table %s\n", len(triggers), table)
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
