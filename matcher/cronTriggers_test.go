package matcher

import (
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/HAL-xyz/zoroaster/tokenapi"
	"github.com/HAL-xyz/zoroaster/trigger"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func init() {
	if config.Zconf.Stage != config.TEST {
		log.Fatal("$STAGE must be TEST to run tests")
	}
	log.SetLevel(log.DebugLevel)
}

func newDateWithTime(day, month, year, hour, min int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC)
}

func TestCronExecutor(t *testing.T) {

	// clear up the database
	err := psqlClient.TruncateTables([]string{"triggers", "matches"})
	assert.NoError(t, err)

	// load a User
	userUUID, err := psqlClient.SaveUser(100, 0)
	assert.NoError(t, err)

	tg1 := `
		{
		  "TriggerName":"Calls SYMBOL on DAI",
		  "TriggerType":"CronTrigger",
		  "ContractAdd":"0x6b175474e89094c44da98b954eedeac495271d0f",
		  "ContractABI":"[{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]",
		  "FunctionName": "symbol",
		  "Inputs": [],
		  "CronJob": {
			"Rule": "*/5 * * * *",
			"Timezone": "-0000"
		  }
		}
	`

	tg2 := `
		{
		  "TriggerName":"Calls SYMBOL on UNI",
		  "TriggerType":"CronTrigger",
		  "ContractAdd":"0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
		  "ContractABI":"[{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]",
		  "FunctionName": "symbol",
		  "Inputs": [],
		  "CronJob": {
			"Rule": "10 15 * * *",
			"Timezone": "-0000"
		  }
		}
	`

	uuid1, err := psqlClient.SaveTrigger(tg1, true, false, userUUID, "1_eth_mainnet")
	uuid2, err := psqlClient.SaveTrigger(tg2, true, false, userUUID, "1_eth_mainnet")

	var api = tokenapi.New(tokenapi.NewZRPC(config.Zconf.EthNode, "mainnet test client"))
	ch := make(chan trigger.IMatch, 2)
	assert.Equal(t, 0, len(ch))

	// Exec at 15:00
	// default date is 1/1/2000 00:00:00 so only the */5 trigger should fire
	CronExecutor(psqlClient, newDateWithTime(1, 1, 2000, 15, 00), api, ch)

	tgs, err := psqlClient.LoadTriggersFromDB(trigger.CronT)
	assert.NoError(t, err)

	tm := tgsToMap(tgs)
	assert.Len(t, tm, 2)
	assert.Equal(t, "2000-01-01 15:00:00 +0000 UTC", tm[uuid1].LastFired.String()) // updated
	assert.Equal(t, "2000-01-01 00:00:00 +0000 UTC", tm[uuid2].LastFired.String()) // not updated

	assert.Equal(t, 1, len(ch))
	m1 := <-ch
	assert.Equal(t, "DAI", m1.ToTemplateMatch().Contract.ReturnedValues[0])

	// Exec at 15:10
	// now both should be executed
	CronExecutor(psqlClient, newDateWithTime(1, 1, 2000, 15, 10), api, ch)

	tgs, err = psqlClient.LoadTriggersFromDB(trigger.CronT)
	assert.NoError(t, err)

	tm = tgsToMap(tgs)
	assert.Len(t, tm, 2)
	assert.Equal(t, "2000-01-01 15:10:00 +0000 UTC", tm[uuid1].LastFired.String()) // updated
	assert.Equal(t, "2000-01-01 15:10:00 +0000 UTC", tm[uuid2].LastFired.String()) // updated

	assert.Equal(t, 2, len(ch))
	m2, m1 := <-ch, <-ch
	assert.Equal(t, "DAI", m1.ToTemplateMatch().Contract.ReturnedValues[0])
	assert.Equal(t, "UNI", m2.ToTemplateMatch().Contract.ReturnedValues[0])

	// after 5 minutes, only the */5 will fire again
	CronExecutor(psqlClient, newDateWithTime(1, 1, 2000, 15, 15), api, ch)

	tgs, err = psqlClient.LoadTriggersFromDB(trigger.CronT)
	assert.NoError(t, err)

	tm = tgsToMap(tgs)
	assert.Len(t, tgs, 2)
	assert.Equal(t, "2000-01-01 15:15:00 +0000 UTC", tm[uuid1].LastFired.String()) // updated
	assert.Equal(t, "2000-01-01 15:10:00 +0000 UTC", tm[uuid2].LastFired.String()) // not updated

	assert.Equal(t, 1, len(ch))
	m1 = <-ch
	assert.Equal(t, "DAI", m1.ToTemplateMatch().Contract.ReturnedValues[0])
}

func tgsToMap(tgs []*trigger.Trigger) map[string]*trigger.Trigger {
	m := make(map[string]*trigger.Trigger)
	for _, tg := range tgs {
		m[tg.TriggerUUID] = tg
	}
	return m
}

func TestShouldFire(t *testing.T) {

	tg1 := &trigger.Trigger{
		CronJob:   trigger.CronJob{Rule: "* * * * *", Timezone: "+0000"}, // every minute
		LastFired: newDateWithTime(1, 1, 2000, 10, 0),                    // next time: 10:01
	}
	assert.False(t, shouldFire(tg1, newDateWithTime(1, 1, 2000, 10, 0))) // before next time
	assert.True(t, shouldFire(tg1, newDateWithTime(1, 1, 2000, 10, 1)))  // eq next time
	assert.True(t, shouldFire(tg1, newDateWithTime(1, 1, 2001, 10, 0)))  // after next time

	tg2 := &trigger.Trigger{
		CronJob:   trigger.CronJob{Rule: "10 15 * * *", Timezone: "+0000"}, // at 15:10 every day
		LastFired: newDateWithTime(1, 1, 2000, 10, 0),                      // next time: 15:10
	}

	assert.False(t, shouldFire(tg2, newDateWithTime(1, 1, 2000, 15, 0))) // before next time
	assert.True(t, shouldFire(tg2, newDateWithTime(1, 1, 2000, 15, 10))) // eq next time
	assert.True(t, shouldFire(tg2, newDateWithTime(1, 1, 2000, 15, 30))) // after next time

	tg3 := &trigger.Trigger{
		CronJob:   trigger.CronJob{Rule: "0 10,15,19 * * *", Timezone: "+0000"}, // at 10, 15 and 19 every day
		LastFired: newDateWithTime(1, 1, 2000, 10, 30),                          // next time: 15:00
	}
	assert.False(t, shouldFire(tg3, newDateWithTime(1, 1, 2000, 11, 0))) // before next time
	assert.True(t, shouldFire(tg3, newDateWithTime(1, 1, 2000, 15, 0)))  // eq next time
	assert.True(t, shouldFire(tg3, newDateWithTime(1, 1, 2000, 15, 10))) // after next time

	tg4 := &trigger.Trigger{
		CronJob:   trigger.CronJob{Rule: "0 0-5,10 * * *", Timezone: "+0000"}, // every hour between 00 and 5, and at 10
		LastFired: newDateWithTime(1, 1, 2000, 1, 30),                         // next time: 2:00
	}
	assert.False(t, shouldFire(tg4, newDateWithTime(1, 1, 2000, 1, 45))) // before next time
	assert.True(t, shouldFire(tg4, newDateWithTime(1, 1, 2000, 2, 0)))   // eq next time
	assert.True(t, shouldFire(tg4, newDateWithTime(1, 1, 2000, 2, 10)))  // after next time

	tg5 := &trigger.Trigger{
		CronJob:   trigger.CronJob{Rule: "*/5 * * * *", Timezone: "+0000"}, // every 5 minutes
		LastFired: newDateWithTime(1, 1, 2000, 1, 30),                      // next time: 1:35
	}
	assert.False(t, shouldFire(tg5, newDateWithTime(1, 1, 2000, 1, 33))) // before next time
	assert.True(t, shouldFire(tg5, newDateWithTime(1, 1, 2000, 2, 35)))  // eq next time
	assert.True(t, shouldFire(tg5, newDateWithTime(1, 1, 2000, 2, 40)))  // after next time

}

func TestShouldFireWithTimezone(t *testing.T) {

	tg6 := &trigger.Trigger{
		CronJob:   trigger.CronJob{Rule: "10 15 * * *", Timezone: "+0100"}, // at 14:10 every day at +0100
		LastFired: newDateWithTime(1, 1, 2000, 10, 0),                      // next time: 14:10
	}
	assert.False(t, shouldFire(tg6, newDateWithTime(1, 1, 2000, 14, 0))) // before next time
	assert.True(t, shouldFire(tg6, newDateWithTime(1, 1, 2000, 14, 10))) // eq next time already!
	assert.True(t, shouldFire(tg6, newDateWithTime(1, 1, 2000, 15, 30))) // after next time

	tg7 := &trigger.Trigger{
		CronJob:   trigger.CronJob{Rule: "10 23 * * *", Timezone: "-0100"}, // at 00:10 every of the *next* day at -0100
		LastFired: newDateWithTime(1, 1, 2000, 10, 0),                      // next time: 00:10 on Jan 2nd
	}
	assert.False(t, shouldFire(tg7, newDateWithTime(1, 1, 2000, 23, 15))) // still before next time!
	assert.True(t, shouldFire(tg7, newDateWithTime(2, 1, 2000, 00, 10)))  // eq next time
	assert.True(t, shouldFire(tg7, newDateWithTime(2, 1, 2000, 15, 30)))  // after next time
}

func TestTriggersToExec(t *testing.T) {
	tg1 := &trigger.Trigger{
		TriggerUUID: "1",
		CronJob:     trigger.CronJob{Rule: "0 4 * * *", Timezone: "+0200"}, // at 4:00 every day, 2:00 UTC
		LastFired:   newDateWithTime(31, 12, 1999, 2, 0),                   // next time: at 2.00 UTC
	}

	tg2 := &trigger.Trigger{
		TriggerUUID: "2",
		CronJob:     trigger.CronJob{Rule: "10 15 * * *", Timezone: "-0000"}, // at 15:10 every day
		LastFired:   newDateWithTime(1, 1, 2000, 10, 0),                      // next time: 15:10
	}

	tg4 := &trigger.Trigger{
		TriggerUUID: "4",
		CronJob:     trigger.CronJob{Rule: "0 0-5,10 * * *", Timezone: "-0000"}, // every hour between 00 and 5, and at 10
		LastFired:   newDateWithTime(1, 1, 2000, 1, 30),                         // next time: 2:00
	}

	tg5 := &trigger.Trigger{
		TriggerUUID: "5",
		CronJob:     trigger.CronJob{Rule: "*/5 * * * *", Timezone: "-0000"}, // every 5 minutes
		LastFired:   newDateWithTime(1, 1, 2000, 1, 30),                      // next time: 1:35
	}

	toRun := filterTgsToRun([]*trigger.Trigger{tg1, tg2, tg4, tg5}, newDateWithTime(1, 1, 2000, 2, 0))
	assert.Len(t, toRun, 3)
	assert.Equal(t, "1", toRun[0].TriggerUUID)
	assert.Equal(t, "4", toRun[1].TriggerUUID)
	assert.Equal(t, "5", toRun[2].TriggerUUID)
}
