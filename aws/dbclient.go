package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"zoroaster/triggers"
)

var db *sql.DB

func main() {
	InitDB()
	triggers, err := loadTriggersFromDB()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("I got %d triggers\n", len(triggers))
	for _, v := range triggers {
		fmt.Println(v.TriggerType)
		fmt.Println(v.Filters)
	}
}

func loadTriggersFromDB() ([]*trigger.Trigger, error) {
	rows, err := db.Query("SELECT trigger_data FROM trigger1;")
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
	)
	pwd := os.Getenv("ZORO_PASS")
	if pwd == "" {
		log.Fatal("No db password set in local env")
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
