package aws

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"time"
	"zoroaster/config"
	"zoroaster/trigger"
)

var db *sql.DB

type PostgresClient struct {
	conf *config.TriggersDB
}

func (cli PostgresClient) GetSilentButMatchingTriggers(triggerUUIDs []string) []string {
	q := fmt.Sprintf(
		`SELECT uuid FROM %s
			WHERE uuid = ANY($1) 
			AND triggered = false`, cli.conf.TableTriggers)

	rows, err := db.Query(q, pq.Array(triggerUUIDs))
	if err != nil {
		log.Error(err)
		return []string{}
	}
	defer rows.Close()

	uuidsRet := make([]string, 0)
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		if err != nil {
			log.Error(err)
		}
		uuidsRet = append(uuidsRet, uuid)
	}
	if err = rows.Err(); err != nil {
		log.Error(err)
	}
	return uuidsRet
}

func (cli PostgresClient) UpdateNonMatchingTriggers(triggerUUIDs []string) {
	q := fmt.Sprintf(
		`UPDATE %s
			SET triggered = false
			WHERE uuid = ANY($1) AND (triggered = true OR triggered IS NULL)`, cli.conf.TableTriggers)

	_, err := db.Exec(q, pq.Array(triggerUUIDs))

	if err != nil {
		log.Errorf("cannot update non-matching triggers: %s", err)
	}
}

func (cli PostgresClient) UpdateMatchingTriggers(triggerUUIDs []string) {
	q := fmt.Sprintf(
		`UPDATE %s
			SET triggered = true
			WHERE uuid = ANY($1) AND (triggered = false OR triggered IS NULL)`, cli.conf.TableTriggers)

	_, err := db.Exec(q, pq.Array(triggerUUIDs))

	if err != nil {
		log.Errorf("cannot update matching triggers: %s", err)
	}
}

func (cli PostgresClient) LogOutcome(outcome *trigger.Outcome, matchUUID string) {
	q := fmt.Sprintf(
		`INSERT INTO %s (
			"match_uuid",
			"payload",
			"outcome",
			"created_at") VALUES ($1::uuid, $2, $3, $4)`, cli.conf.TableOutcomes)

	_, err := db.Exec(q, matchUUID, outcome.Payload, outcome.Outcome, time.Now())
	if err != nil {
		log.Errorf("cannot log outcome: %s", err)
	}
}

func (cli PostgresClient) GetActions(tgUUID string, userUUID string) ([]string, error) {
	q := fmt.Sprintf(
		`SELECT action_data
				FROM %s AS tg_table, %s AS act_table
				WHERE tg_table.user_uuid = $1::uuid
				AND tg_table.uuid = act_table.trigger_uuid
				AND tg_table.uuid = $2::uuid
				AND tg_table.is_active = true`,
		cli.conf.TableTriggers, cli.conf.TableActions)
	rows, err := db.Query(q, userUUID, tgUUID)
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

func (cli PostgresClient) ReadLastBlockProcessed(tgType trigger.TgType) int {
	var blockNo int
	q := fmt.Sprintf("SELECT %s_last_block_processed FROM %s", trigger.TgTypeToPrefix(tgType), cli.conf.TableState)
	err := db.QueryRow(q).Scan(&blockNo)
	if err != nil {
		log.Errorf("cannot read last block processed: %s", err)
	}
	return blockNo
}

func (cli PostgresClient) SetLastBlockProcessed(blockNo int, tgType trigger.TgType) {
	stringTgType := trigger.TgTypeToPrefix(tgType)
	q := fmt.Sprintf(`UPDATE "%s" SET %s_last_block_processed = $1, %s_date = $2`, cli.conf.TableState, stringTgType, stringTgType)
	_, err := db.Exec(q, blockNo, time.Now())
	if err != nil {
		log.Errorf("cannot set last block processed: %s", err)
	}
}

func (cli PostgresClient) LogMatch(match trigger.IMatch) string {
	matchData, err := json.Marshal(match.ToPersistent())
	if err != nil {
		log.Errorf("cannot marshall match into json")
		return ""
	}
	q := fmt.Sprintf(
		`INSERT INTO "%s" (
			"trigger_uuid", "match_data", "created_at")
			VALUES ($1, $2, $3) RETURNING uuid`, cli.conf.TableMatches)
	var lastUUID string
	err = db.QueryRow(q, match.GetTriggerUUID(), matchData, time.Now()).Scan(&lastUUID)
	if err != nil {
		log.Errorf("cannot write log iMatch: %s", err)
	}
	return lastUUID
}

func (cli PostgresClient) LoadTriggersFromDB(tgType trigger.TgType) ([]*trigger.Trigger, error) {
	q := fmt.Sprintf(
		`SELECT uuid, trigger_data, user_uuid
				FROM %s AS t
				WHERE (t.trigger_data ->> 'TriggerType')::text = '%s'
				AND t.is_active = true`, cli.conf.TableTriggers, trigger.TgTypeToString(tgType))
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	triggers := make([]*trigger.Trigger, 0)
	for rows.Next() {
		var triggerUUID, userUUID string
		var tg string
		err = rows.Scan(&triggerUUID, &tg, &userUUID)
		if err != nil {
			return nil, err
		}
		trig, err := trigger.NewTriggerFromJson(tg)
		if err != nil {
			log.Debugf("(trigger uuid %s): %v", triggerUUID, err)
			return nil, err
		} else {
			trig.TriggerUUID, trig.UserUUID = triggerUUID, userUUID
			triggers = append(triggers, trig)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return triggers, nil
}

func (cli *PostgresClient) InitDB(c *config.ZConfiguration) {
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

	cli.conf = &c.TriggersDB
}

func (cli PostgresClient) Close() {
	err := db.Close()
	if err != nil {
		log.Error(err)
	}
}
