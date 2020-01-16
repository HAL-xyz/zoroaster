package aws

import (
	"fmt"
	"time"
)

// Helper functions used for tests only

func (cli PostgresClient) ReadString(query string) (string, error) {
	var output string
	err := db.QueryRow(query).Scan(&output)
	if err != nil {
		return "", fmt.Errorf("cannot read string: %s", err)
	}
	return output, nil
}

func (cli PostgresClient) SaveUser() (string, error) {
	q := fmt.Sprintf(
		`INSERT INTO users (
			"display_name", 
			"email", 
			"actions_monthly_cap",
			"user_type",
			"created_at",
			"counter_current_month") VALUES ($1, $2, $3, $4, $5, $6) RETURNING uuid`)
	var lastUUID string
	err := db.QueryRow(q, "batman", "email@lol.com", 1, "admin", time.Now(), 1).Scan(&lastUUID)
	return lastUUID, err
}

func (cli PostgresClient) SaveTrigger(triggerData string, isActive, triggered bool, userId string) (string, error) {
	q := fmt.Sprintf(
		`INSERT INTO triggers (
			"trigger_data", 
			"is_active", 
			"created_at",
			"updated_at",
			"triggered",
			"user_uuid") VALUES ($1, $2, $3, $4, $5, $6) RETURNING uuid`)
	var lastUUID string
	err := db.QueryRow(q, triggerData, isActive, time.Now(), time.Now(), triggered, userId).Scan(&lastUUID)
	return lastUUID, err
}

func (cli PostgresClient) SaveAction(triggerUUID string) (string, error) {
	q := fmt.Sprintf(
		`INSERT INTO actions (
			"action_data", 
			"is_active", 
			"trigger_uuid",
			"created_at",
			"updated_at") VALUES ($1, $2, $3, $4, $5) RETURNING uuid`)
	var lastUUID string
	actionData := `{
  "ActionType": "webhook_post",
  "Attributes": {
    "URI": "https://webhook.site/3e94a980-cc28-4fb3-8733-8e398e20c066"
  }
}`
	err := db.QueryRow(q, actionData, true, triggerUUID, time.Now(), time.Now()).Scan(&lastUUID)
	return lastUUID, err

}
