package aws

import (
	"zoroaster/config"
	"zoroaster/trigger"
)

type IDB interface {
	InitDB(c *config.ZConfiguration)

	Close()

	LoadTriggersFromDB(table string, watOrWac string) ([]*trigger.Trigger, error)

	LogOutcome(table string, outcome *trigger.Outcome, matchId int)

	GetActions(table string, tgId int, userId int) ([]string, error)

	ReadLastBlockProcessed(table string, watOrWac string) int

	SetLastBlockProcessed(table string, blockNo int, watOrWac string)

	LogTxMatch(table string, match trigger.TxMatch) int

	LogCnMatch(table string, match trigger.CnMatch) int

	UpdateMatchingTriggers(table string, triggerIds []int)

	UpdateNonMatchingTriggers(table string, triggerIds []int)
}
