package aws

import (
	"zoroaster/config"
	"zoroaster/trigger"
)

type IDB interface {
	InitDB(c *config.ZConfiguration)

	Close()

	LoadTriggersFromDB(tgType trigger.TgType) ([]*trigger.Trigger, error)

	LogOutcome(outcome *trigger.Outcome, matchUUID string)

	GetActions(tgUUID string, userUUID string) ([]string, error)

	ReadLastBlockProcessed(tgType trigger.TgType) int

	SetLastBlockProcessed(blockNo int, tgType trigger.TgType) error

	LogMatch(match trigger.IMatch) string

	UpdateMatchingTriggers(triggerIds []string)

	UpdateNonMatchingTriggers(triggerIds []string)

	GetSilentButMatchingTriggers(triggerUUIDs []string) []string
}
