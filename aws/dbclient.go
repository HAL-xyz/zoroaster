package aws

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/onrik/ethrpc"
	"log"
	"time"
	"zoroaster/config"
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

func InitDB(c *config.ZConfiguration) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.TriggersDB.Endpoint, c.TriggersDB.Port, c.TriggersDB.User, c.TriggersDB.Password, c.TriggersDB.Name)

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
