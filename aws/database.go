package aws

import "zoroaster/trigger"

type IDB interface {

	LoadTriggersFromDB(table string) ([]*trigger.Trigger, error)

}

