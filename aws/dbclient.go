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

func (cli PostgresClient) UpdateNonMatchingTriggers(triggerIds []int) {
	q := fmt.Sprintf(
		`UPDATE %s
			SET triggered = false
			WHERE id = ANY($1) AND (triggered = true OR triggered IS NULL)`, cli.conf.TableTriggers)

	_, err := db.Exec(q, pq.Array(triggerIds))

	if err != nil {
		log.Errorf("cannot update non-matching triggers: %s", err)
	}
}

func (cli PostgresClient) UpdateMatchingTriggers(triggerIds []int) {
	q := fmt.Sprintf(
		`UPDATE %s
			SET triggered = true
			WHERE id = ANY($1) AND (triggered = false OR triggered IS NULL)`, cli.conf.TableTriggers)

	_, err := db.Exec(q, pq.Array(triggerIds))

	if err != nil {
		log.Errorf("cannot update matching triggers: %s", err)
	}
}

func (cli PostgresClient) LogOutcome(outcome *trigger.Outcome, matchId int) {
	q := fmt.Sprintf(
		`INSERT INTO %s (
			"match_id",
			"payload",
			"outcome",
			"created_at") VALUES ($1, $2, $3, $4)`, cli.conf.TableOutcomes)

	_, err := db.Exec(q, matchId, outcome.Payload, outcome.Outcome, time.Now())
	if err != nil {
		log.Errorf("cannot log outcome: %s", err)
	}
}

func (cli PostgresClient) GetActions(tgId int, userId int) ([]string, error) {
	q := fmt.Sprintf(
		`SELECT action_data
				FROM %s AS tg_table, %s AS act_table
				WHERE tg_table.user_id = $1
				AND tg_table.id = act_table.trigger_id
				AND tg_table.id = $2
				AND tg_table.is_active = true`,
		cli.conf.TableTriggers, cli.conf.TableActions)
	rows, err := db.Query(q, userId, tgId)
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

func (cli PostgresClient) ReadLastBlockProcessed(watOrWac string) int {
	var blockNo int
	q := fmt.Sprintf("SELECT %s_last_block_processed FROM %s", watOrWac, cli.conf.TableStats)
	err := db.QueryRow(q).Scan(&blockNo)
	if err != nil {
		log.Errorf("cannot read last block processed: %s", err)
	}
	return blockNo
}

func (cli PostgresClient) SetLastBlockProcessed(blockNo int, watOrWac string) {
	q := fmt.Sprintf(`UPDATE "%s" SET %s_last_block_processed = $1, %s_date = $2`, cli.conf.TableStats, watOrWac, watOrWac)
	_, err := db.Exec(q, blockNo, time.Now())
	if err != nil {
		log.Errorf("cannot set last block processed: %s", err)
	}
}

func (cli PostgresClient) LogMatch(match trigger.IMatch) int {
	matchData, err := json.Marshal(match.ToPersistent())
	if err != nil {
		log.Errorf("cannot marshall match into json")
		return -1
	}
	q := fmt.Sprintf(
		`INSERT INTO "%s" (
			"trigger_id", "match_data", "created_at")
			VALUES ($1, $2, $3) RETURNING id`, cli.conf.TableMatches)
	var lastId int
	err = db.QueryRow(q, match.GetTriggerId(), matchData, time.Now()).Scan(&lastId)
	if err != nil {
		log.Errorf("cannot write log iMatch: %s", err)
	}
	return lastId
}

// TODO: need to make `watOrWac` its own type cuz I never remember what to plug in there otherwise
func (cli PostgresClient) LoadTriggersFromDB(watOrWac string) ([]*trigger.Trigger, error) {
	q := fmt.Sprintf(
		`SELECT id, trigger_data, user_id
				FROM %s AS t
				WHERE (t.trigger_data ->> 'TriggerType')::text = '%s'
				AND t.is_active = true`, cli.conf.TableTriggers, watOrWac)
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
