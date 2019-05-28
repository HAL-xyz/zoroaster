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

func ReadLastBlockProcessed(table string) int {
	var blockNo int
	q := fmt.Sprintf("SELECT last_block_processed FROM %s", table)
	err := db.QueryRow(q).Scan(&blockNo)
	if err != nil {
		log.Printf("ERROR: Cannot read last block processed: %s", err)
	}
	return blockNo
}

func SetLastBlockProcessed(table string, blockNo int) {
	q := fmt.Sprintf(`UPDATE "%s" SET last_block_processed = $1, date = $2`, table)
	_, err := db.Exec(q, blockNo, time.Now())
	if err != nil {
		log.Printf("ERROR: Cannot set last block processed: %s", err)
	}
}

func LogMatch(tg *trigger.Trigger, tx *ethrpc.Transaction, logTable string) {
	q := fmt.Sprintf(`INSERT INTO "%s" ("date", "trigger_id", "block_no", "tx_hash") VALUES($1, $2, $3, $4)`, logTable)
	_, err := db.Exec(q, time.Now(), tg.TriggerId, *tx.BlockNumber, tx.Hash)
	if err != nil {
		log.Printf("WARN: Cannot write trigger log match: %s", err)
	}
}

func LoadTriggersFromDB(table string) ([]*trigger.Trigger, error) {
	q := fmt.Sprintf("SELECT trigger_data FROM %s", table)
	rows, err := db.Query(q)
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
