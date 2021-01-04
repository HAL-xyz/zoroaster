package db

import (
	"fmt"
	"github.com/HAL-xyz/zoroaster/aws"
	"github.com/HAL-xyz/zoroaster/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var psqlClient = aws.NewPostgresClient(config.Zconf)

func init() {
	if config.Zconf.Stage != config.TEST {
		logrus.Fatal("$STAGE must be TEST to run db tests")
	}
}

func TestMonthlyRun(t *testing.T) {

	defer psqlClient.Close()

	// load a User
	userUUID, err := psqlClient.SaveUser(100, 25)
	assert.NoError(t, err)

	// set current month state = 11
	// monthlyDbUpdate(7) won't update anything
	err = psqlClient.SetString(fmt.Sprintf(`UPDATE state SET current_month = 11`))
	assert.NoError(t, err)

	err = monthlyDbUpdate(psqlClient, 11)
	assert.NoError(t, err)

	currMonth, err := psqlClient.ReadString(fmt.Sprintf(`SELECT current_month FROM state`))
	assert.NoError(t, err)
	assert.Equal(t, "11", currMonth)

	// now let's run it when we're in December
	err = monthlyDbUpdate(psqlClient, 12)
	assert.NoError(t, err)

	currMonth, err = psqlClient.ReadString(fmt.Sprintf(`SELECT current_month FROM state`))
	assert.NoError(t, err)
	assert.Equal(t, "12", currMonth)

	userCounter, err := psqlClient.ReadString(fmt.Sprintf(`SELECT counter_current_month FROM users WHERE uuid = '%s'`, userUUID))
	assert.NoError(t, err)
	assert.Equal(t, "0", userCounter)

	// January, it's been a year!
	err = monthlyDbUpdate(psqlClient, 1)
	assert.NoError(t, err)

	currMonth, err = psqlClient.ReadString(fmt.Sprintf(`SELECT current_month FROM state`))
	assert.NoError(t, err)
	assert.Equal(t, "1", currMonth)

	userCounter, err = psqlClient.ReadString(fmt.Sprintf(`SELECT counter_current_month FROM users WHERE uuid = '%s'`, userUUID))
	assert.NoError(t, err)
	assert.Equal(t, "0", userCounter)

	// make sure January only updates once
	err = psqlClient.SetString(fmt.Sprintf(`UPDATE users SET counter_current_month = 50 WHERE uuid = '%s'`, userUUID))
	userCounter, err = psqlClient.ReadString(fmt.Sprintf(`SELECT counter_current_month FROM users WHERE uuid = '%s'`, userUUID))
	assert.NoError(t, err)
	assert.Equal(t, "50", userCounter)

	err = monthlyDbUpdate(psqlClient, 1)
	assert.NoError(t, err)

	currMonth, err = psqlClient.ReadString(fmt.Sprintf(`SELECT current_month FROM state`))
	assert.NoError(t, err)
	assert.Equal(t, "1", currMonth)

	userCounter, err = psqlClient.ReadString(fmt.Sprintf(`SELECT counter_current_month FROM users WHERE uuid = '%s'`, userUUID))
	assert.NoError(t, err)
	assert.Equal(t, "50", userCounter)

}
