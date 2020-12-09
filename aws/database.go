package aws

import (
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/trigger"
	"time"
)

type IDB interface {
	initDB(c *config.ZConfiguration)

	Close()

	LoadTriggersFromDB(tgType trigger.TgType) ([]*trigger.Trigger, error)

	LogOutcome(outcome *trigger.Outcome, matchUUID string) error

	GetActions(tgUUID string, userUUID string) ([]string, error)

	ReadLastBlockProcessed(tgType trigger.TgType) (int, error)

	SetLastBlockProcessed(blockNo int, tgType trigger.TgType) error

	LogMatch(match trigger.IMatch) (string, error)

	UpdateMatchingTriggers(triggerIds []string)

	UpdateNonMatchingTriggers(triggerIds []string)

	GetSilentButMatchingTriggers(triggerUUIDs []string) ([]string, error)

	LogAnalytics(tgType trigger.TgType, blockNo, triggersNo, blockTime int, start, end time.Time) error

	ReadSavedMonth() (int, error)

	UpdateSavedMonth(newMonth int) error
}
