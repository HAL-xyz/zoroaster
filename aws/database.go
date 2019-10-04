package aws

import (
	"zoroaster/config"
	"zoroaster/trigger"
)

type IDB interface {
	InitDB(c *config.ZConfiguration)

	Close()

	LoadTriggersFromDB(watOrWac string) ([]*trigger.Trigger, error)

	LogOutcome(outcome *trigger.Outcome, matchId int, watOrWac string)

	GetActions(tgId int, userId int) ([]string, error)

	ReadLastBlockProcessed(watOrWac string) int

	SetLastBlockProcessed(blockNo int, watOrWac string)

	LogMatch(match trigger.IMatch) int

	UpdateMatchingTriggers(triggerIds []int)

	UpdateNonMatchingTriggers(triggerIds []int)
}
