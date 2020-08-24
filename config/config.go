package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type ZConfiguration struct {
	Stage                 Stage
	ConfigFile            string
	EthNode               string // the main eth node
	RinkebyNode           string // Rinkeby network, used for tests
	LogsPath              string
	LogsFile              string
	Database              ZoroDB
	BlocksDelay           int
	TwitterConsumerKey    string
	TwitterConsumerSecret string
}

type ZoroDB struct {
	TableTriggers string
	TableMatches  string
	TableOutcomes string
	TableState    string
	TableActions  string
	TableUsers    string
	Host          string
	User          string
	Name          string
	Port          int
	Password      string
}

type Stage int

const (
	TEST Stage = iota
	DEV
	PROD
)

func (s Stage) String() string {
	return [...]string{"TEST", "DEV", "PROD"}[s]
}

// ENV variables
const (
	logsPath              = "LOGS_PATH"
	blocksDelay           = "BLOCKS_DELAY"
	dbHost                = "DB_HOST"
	dbName                = "DB_NAME"
	dbUsr                 = "DB_USR"
	dbPwd                 = "DB_PWD"
	ethNode               = "ETH_NODE"
	rinkebyNode           = "RINKEBY_NODE"
	twitterConsumerKey    = "TWITTER_CONSUMER_KEY"
	twitterConsumerSecret = "TWITTER_CONSUMER_SECRET"
)

// DB tables
const (
	dbPort        = 5432
	tableTriggers = "triggers"
	tableMatches  = "matches"
	tableOutcomes = "outcomes"
	tableState    = "state"
	tableActions  = "actions"
	tableUsers    = "users"
)

func NewConfig() *ZConfiguration {

	zconfig := ZConfiguration{}

	stage := os.Getenv("STAGE")
	switch stage {
	case "TEST":
		zconfig.Stage = TEST
	case "DEV":
		zconfig.Stage = DEV
	case "PROD":
		zconfig.Stage = PROD
	default:
		log.Fatal("local env STAGE must be TEST, DEV or PROD")
	}

	zconfig.Database.TableTriggers = tableTriggers
	zconfig.Database.TableMatches = tableMatches
	zconfig.Database.TableOutcomes = tableOutcomes
	zconfig.Database.TableState = tableState
	zconfig.Database.TableActions = tableActions
	zconfig.Database.TableUsers = tableUsers
	zconfig.Database.Port = dbPort

	zconfig.LogsFile = os.Getenv(logsPath)
	if zconfig.LogsFile == "" {
		log.Fatal("no logs path set in local env ", logsPath)
	}
	zconfig.LogsFile = fmt.Sprintf("%s/%s.log", zconfig.LogsPath, zconfig.Stage)

	delay := os.Getenv(blocksDelay)
	intDelay, err := strconv.Atoi(delay)
	if delay == "" || err != nil {
		log.Fatalf("cannot use %s as block delay", delay)
	}
	zconfig.BlocksDelay = intDelay

	zconfig.Database.Host = os.Getenv(dbHost)
	if zconfig.Database.Host == "" {
		log.Fatal("no db host set in local env ", dbHost)
	}

	zconfig.Database.Name = os.Getenv(dbName)
	if zconfig.Database.Name == "" {
		log.Fatal("no db name set in local env ", dbName)
	}

	zconfig.Database.User = os.Getenv(dbUsr)
	if zconfig.Database.User == "" {
		log.Fatal("no db user set in local env ", dbUsr)
	}

	zconfig.Database.Password = os.Getenv(dbPwd)
	if zconfig.Database.Password == "" {
		log.Fatal("no db password set in local env ", dbPwd)
	}

	zconfig.EthNode = os.Getenv(ethNode)
	if zconfig.EthNode == "" {
		log.Fatal("no eth node set in local env ", ethNode)
	}

	// Rinkeby node is only required for tests
	zconfig.RinkebyNode = os.Getenv(rinkebyNode)
	if zconfig.Stage == TEST && zconfig.EthNode == "" {
		log.Error("no Rinkeby node set in local env ", rinkebyNode)
	}

	zconfig.TwitterConsumerKey = os.Getenv(twitterConsumerKey)
	if zconfig.TwitterConsumerKey == "" {
		log.Fatal("no twitter consumer key set in local env ", twitterConsumerKey)
	}

	zconfig.TwitterConsumerSecret = os.Getenv(twitterConsumerSecret)
	if zconfig.TwitterConsumerSecret == "" {
		log.Fatal("no twitter consumer secret set in local env ", twitterConsumerSecret)
	}

	return &zconfig
}
