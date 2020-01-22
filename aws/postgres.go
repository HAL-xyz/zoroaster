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
	conf *config.ZoroDB
}

func (cli PostgresClient) LogAnalytics(tgType trigger.TgType, blockNo, triggersNo, blockTime int, start, end time.Time) error {
	q := fmt.Sprintf(
		`INSERT INTO analytics (
			"type", 
			"block_no", 
			"no_triggers",
			"start_time",
			"end_time",
			"duration",
			"block_time") VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	_, err := db.Exec(q, trigger.TgTypeToPrefix(tgType), blockNo, triggersNo, start, end, end.Sub(start).Seconds(), time.Unix(int64(blockTime), 0))
	return err
}

func (cli PostgresClient) GetSilentButMatchingTriggers(triggerUUIDs []string) ([]string, error) {
	q := fmt.Sprintf(
		`SELECT uuid FROM %s
			WHERE uuid = ANY($1) 
			AND triggered = false`, cli.conf.TableTriggers)

	rows, err := db.Query(q, pq.Array(triggerUUIDs))
	if err != nil {
		return []string{}, err
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
	return uuidsRet, nil
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

func (cli PostgresClient) LogOutcome(outcome *trigger.Outcome, matchUUID string) error {
	q := fmt.Sprintf(
		`INSERT INTO %s (
			"match_uuid",
			"payload_data",
			"outcome_data",
			"created_at") VALUES ($1::uuid, $2, $3, $4)`, cli.conf.TableOutcomes)

	_, err := db.Exec(q, matchUUID, outcome.Payload, outcome.Outcome, time.Now())
	if err != nil {
		return fmt.Errorf("cannot log outcome with payload: %s; outcome: %s; error: %s", outcome.Payload, outcome.Outcome, err)
	}
	return nil
}

func (cli PostgresClient) GetActions(tgUUID string, userUUID string) ([]string, error) {
	q := fmt.Sprintf(
		`SELECT action_data
				FROM %s AS tg_table, %s AS act_table
				WHERE tg_table.user_uuid = $1::uuid
				AND tg_table.uuid = act_table.trigger_uuid
				AND tg_table.uuid = $2::uuid
				AND act_table.is_active = true`,
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

func (cli PostgresClient) ReadLastBlockProcessed(tgType trigger.TgType) (int, error) {
	var blockNo int
	q := fmt.Sprintf("SELECT %s_last_block_processed FROM %s", trigger.TgTypeToPrefix(tgType), cli.conf.TableState)
	err := db.QueryRow(q).Scan(&blockNo)
	if err != nil {
		return 0, fmt.Errorf("cannot read last block processed: %s", err)
	}
	return blockNo, nil
}

func (cli PostgresClient) SetLastBlockProcessed(blockNo int, tgType trigger.TgType) error {
	stringTgType := trigger.TgTypeToPrefix(tgType)
	q := fmt.Sprintf(`UPDATE "%s" SET %s_last_block_processed = $1, %s_date = $2`, cli.conf.TableState, stringTgType, stringTgType)
	_, err := db.Exec(q, blockNo, time.Now())
	if err != nil {
		return fmt.Errorf("cannot set last block processed: %s", err)
	}
	return nil
}

func (cli PostgresClient) LogMatch(match trigger.IMatch) (string, error) {
	matchData, err := json.Marshal(match.ToPersistent())
	if err != nil {
		return "", err
	}
	q := fmt.Sprintf(
		`INSERT INTO "%s" (
			"trigger_uuid", "match_data", "created_at")
			VALUES ($1, $2, $3) RETURNING uuid`, cli.conf.TableMatches)
	var lastUUID string
	err = db.QueryRow(q, match.GetTriggerUUID(), matchData, time.Now()).Scan(&lastUUID)
	if err != nil {
		return "", err
	}
	// also update user's counter
	upQ := fmt.Sprintf(`UPDATE "%s"
                SET counter_current_month = counter_current_month + 1 
				WHERE uuid = '%s' `, cli.conf.TableUsers, match.GetUserUUID())
	_, err = db.Exec(upQ)
	if err != nil {
		return "", err
	}
	return lastUUID, nil
}

func (cli PostgresClient) LoadTriggersFromDB(tgType trigger.TgType) ([]*trigger.Trigger, error) {
	q := fmt.Sprintf(
		`SELECT tg_table.uuid, trigger_data, user_uuid
				FROM %s AS tg_table, %s AS usr_table
				WHERE (tg_table.trigger_data ->> 'TriggerType')::text = '%s'
				AND tg_table.user_uuid = usr_table.uuid
				AND counter_current_month < actions_monthly_cap
				AND tg_table.is_active = true`, cli.conf.TableTriggers, cli.conf.TableUsers, trigger.TgTypeToString(tgType))
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
			log.Warnf("trigger uuid %s: %v", triggerUUID, err)
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
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Name)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(psqlInfo)
		log.Fatal("cannot connect to the DB -> ", err)
	}

	cli.conf = &c.Database
}

func (cli PostgresClient) Close() {
	err := db.Close()
	if err != nil {
		log.Error(err)
	}
}
