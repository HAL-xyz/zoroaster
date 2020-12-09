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
	_ = userUUID

	// set current month state = 7
	// monthlyDbUpdate(7) won't update anything
	err = psqlClient.SetString(fmt.Sprintf(`UPDATE state SET current_month = 7`))
	assert.NoError(t, err)

	err = monthlyDbUpdate(psqlClient, 7)
	assert.NoError(t, err)

	currMonth, err := psqlClient.ReadString(fmt.Sprintf(`SELECT current_month FROM state`))
	assert.NoError(t, err)
	assert.Equal(t, "7", currMonth)

	// now let's run it when we're in August
	err = monthlyDbUpdate(psqlClient, 8)
	assert.NoError(t, err)

	currMonth, err = psqlClient.ReadString(fmt.Sprintf(`SELECT current_month FROM state`))
	assert.NoError(t, err)
	assert.Equal(t, "8", currMonth)

	userCounter, err := psqlClient.ReadString(fmt.Sprintf(`SELECT counter_current_month FROM users WHERE uuid = '%s'`, userUUID))
	assert.NoError(t, err)
	assert.Equal(t, "0", userCounter)
}
