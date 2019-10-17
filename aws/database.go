package aws

import (
	"zoroaster/config"
	"zoroaster/trigger"
)

type IDB interface {
	InitDB(c *config.ZConfiguration)

	Close()

	LoadTriggersFromDB(watOrWac string) ([]*trigger.Trigger, error)

	LogOutcome(outcome *trigger.Outcome, matchUUID string)

	GetActions(tgUUID string, userUUID string) ([]string, error)

	ReadLastBlockProcessed(watOrWac string) int

	SetLastBlockProcessed(blockNo int, watOrWac string)

	LogMatch(match trigger.IMatch) string

	UpdateMatchingTriggers(triggerIds []string)

	UpdateNonMatchingTriggers(triggerIds []string)

	GetSilentButMatchingTriggers(triggerUUIDs []string) []string
}
