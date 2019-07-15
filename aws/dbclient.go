package aws

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/config"
	"zoroaster/trigger"
)

var db *sql.DB

type PostgresClient struct{}

func (cli PostgresClient) LogOutcome(table string, outcome *trigger.Outcome, matchId int) {
	q := fmt.Sprintf(
		`INSERT INTO %s (
			"match_id",
			"payload",
			"outcome",
			"timestamp") VALUES ($1, $2, $3, $4)`, table)

	_, err := db.Exec(q, matchId, outcome.Payload, outcome.Outcome, time.Now())
	if err != nil {
		log.Errorf("cannot log outcome: %s", err)
	}
}

func (cli PostgresClient) GetActions(table string, tgId int, userId int) ([]string, error) {
	q := fmt.Sprintf(
		`SELECT action_data
				FROM %s AS t
				WHERE (t.action_data ->> 'TriggerId')::int = $1
				AND (t.action_data ->> 'UserId')::int = $2
				AND t.is_active = true`, table)
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

func (cli PostgresClient) ReadLastBlockProcessed(table string, watOrWac string) int {
	var blockNo int
	q := fmt.Sprintf("SELECT %s_last_block_processed FROM %s", watOrWac, table)
	err := db.QueryRow(q).Scan(&blockNo)
	if err != nil {
		log.Errorf("cannot read last block processed: %s", err)
	}
	return blockNo
}

func (cli PostgresClient) SetLastBlockProcessed(table string, blockNo int, watOrWac string) {
	q := fmt.Sprintf(`UPDATE "%s" SET %s_last_block_processed = $1, %s_date = $2`, table, watOrWac, watOrWac)
	_, err := db.Exec(q, blockNo, time.Now())
	if err != nil {
		log.Errorf("cannot set last block processed: %s", err)
	}
}

func (cli PostgresClient) LogCnMatch(table string, match trigger.CnMatch) int {
	q := fmt.Sprintf(
		`INSERT INTO "%s" (
			"date", "trigger_id", "block_no", "call_value")
			VALUES ($1, $2, $3, $4) RETURNING id`, table)
	var lastId int
	err := db.QueryRow(q, time.Now(), match.TgId, match.BlockNo, match.Value).Scan(&lastId)

	if err != nil {
		log.Errorf("cannot write contract log match: %s", err)
	}
	return lastId
}

func (cli PostgresClient) LogTxMatch(table string, match trigger.TxMatch) int {
	bdate := time.Unix(int64(match.ZTx.BlockTimestamp), 0)
	tx := match.ZTx.Tx
	tg := match.Tg
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
			"user_id",
			"fn_name",
			"fn_args") VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id`, table)

	var lastId int
	err := db.QueryRow(q, time.Now(), tg.TriggerId, *tx.BlockNumber, tx.BlockHash, bdate, tx.Hash, tx.From,
		tx.To, tx.Nonce, tx.Value.String(), tx.GasPrice.String(), tx.Gas, tx.Input, tg.UserId,
		match.ZTx.DecodedFnName, match.ZTx.DecodedFnArgs).Scan(&lastId)
	if err != nil {
		log.Errorf("cannot write transaction log match: %s", err)
	}
	return lastId
}

func (cli PostgresClient) LoadTriggersFromDB(table string, watOrWac string) ([]*trigger.Trigger, error) {
	q := fmt.Sprintf(
		`SELECT id, trigger_data, user_id
				from %s as t
				where (t.trigger_data ->> 'TriggerType')::text = '%s'`, table, watOrWac)
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

func (cli PostgresClient) InitDB(c *config.ZConfiguration) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.TriggersDB.Host, c.TriggersDB.Port, c.TriggersDB.User, c.TriggersDB.Password, c.TriggersDB.Name)

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

func (cli PostgresClient) Close() {
	err := db.Close()
	if err != nil {
		log.Error(err)
	}
}
