package db

import (
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/sirupsen/logrus"
	"time"
)

// Every month we want to update the users table
// to reset all the matches quota.
// We do this by keeping track of what was the last
// new month we've seen, and once every hour we
// compare it with the current month.

func MatchesMonthlyUpdate(idb aws.IDB) {

	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		err := monthlyDbUpdate(idb, time.Now().Month())
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func monthlyDbUpdate(idb aws.IDB, currentMonth time.Month) error {

	savedMonth, err := idb.ReadSavedMonth()
	if err != nil {
		return err
	}

	if savedMonth < int(currentMonth) {
		err = idb.UpdateSavedMonth(int(currentMonth))
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Info("Updated state month to be ", currentMonth)
		logrus.Info("counter_current_month reset for all users successfully")
	}

	return nil
}
