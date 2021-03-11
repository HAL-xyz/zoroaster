package db

import (
	"github.com/sirupsen/logrus"
	"time"
)

// Every month we want to update the users table
// to reset all the matches quota.
// We do this by keeping track of what was the last
// new month we've seen, and once every hour we
// compare it with the current month.

func MatchesMonthlyUpdate(idb IDB) {

	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		err := monthlyDbUpdate(idb, time.Now().Month())
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func monthlyDbUpdate(idb IDB, currentMonth time.Month) error {

	persistedMonth, err := idb.ReadSavedMonth()
	if err != nil {
		return err
	}

	if persistedMonth < int(currentMonth) || (persistedMonth == 12 && currentMonth == 1) {
		err = idb.UpdateSavedMonth(int(currentMonth))
		if err != nil {
			return err
		}
		logrus.Info("Updated state month to be ", currentMonth)
		logrus.Info("counter_current_month reset for all users successfully")
	}

	return nil
}
