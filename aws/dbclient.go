package aws

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/onrik/ethrpc"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/config"
	"zoroaster/triggers"
)

var db *sql.DB

func GetActions(table string, tgId int, userId int) ([]string, error) {
	q := fmt.Sprintf(
		`SELECT action_data
				FROM %s AS t
				WHERE (t.action_data ->> 'TriggerId')::int = $1
				AND (t.action_data ->> 'UserId')::int = $2`, table)
	rows, err := db.Query(q, tgId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actionsRet := make([]string, 0)
	for rows.Next() {
		var action string
		err = rows.Scan(&action)
		if err != nil {
			return nil, err
		}
		actionsRet = append(actionsRet, action)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return actionsRet, nil
}

func ReadLastBlockProcessed(table string) int {
	var blockNo int
	q := fmt.Sprintf("SELECT last_block_processed FROM %s", table)
	err := db.QueryRow(q).Scan(&blockNo)
	if err != nil {
		log.Errorf("cannot read last block processed: %s", err)
	}
	return blockNo
}

func SetLastBlockProcessed(table string, blockNo int) {
	q := fmt.Sprintf(`UPDATE "%s" SET last_block_processed = $1, date = $2`, table)
	_, err := db.Exec(q, blockNo, time.Now())
	if err != nil {
		log.Errorf("cannot set last block processed: %s", err)
	}
}

func LogMatch(table string, tg *trigger.Trigger, tx *ethrpc.Transaction, blockTimestamp int) {
	bdate := time.Unix(int64(blockTimestamp), 0)
	q := fmt.Sprintf(
		`INSERT INTO "%s" (
			"date",
			"trigger_id",
			"block_no",
			"block_hash",
			"block_time",
			"tx_hash",
			"from",
			"to",
			"nonce",
			"value",
			"gas_price",
			"gas",
			"data",
			"user_id") VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`, table)
	_, err := db.Exec(q, time.Now(), tg.TriggerId, *tx.BlockNumber, tx.BlockHash, bdate, tx.Hash, tx.From, tx.To, tx.Nonce, tx.Value.String(), tx.GasPrice.String(), tx.Gas, tx.Input, tg.UserId)
	if err != nil {
		log.Errorf("cannot write trigger log match: %s", err)
	}
}

func LoadTriggersFromDB(table string) ([]*trigger.Trigger, error) {
	q := fmt.Sprintf("SELECT id, trigger_data, user_id FROM %s", table)
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	triggers := make([]*trigger.Trigger, 0)
	for rows.Next() {
		var triggerId, userId int
		var tg string
		err = rows.Scan(&triggerId, &tg, &userId)
		if err != nil {
			return nil, err
		}
		trig, err := trigger.NewTriggerFromJson(tg)
		if err != nil {
			log.Debugf("(trigger id %d): %v", triggerId, err)
			return nil, err
		} else {
			trig.TriggerId, trig.UserId = triggerId, userId
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
		log.Fatal("cannot connect to the DB -> ", err)
	}
}
