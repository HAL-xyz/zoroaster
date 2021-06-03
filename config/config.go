package config

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var Zconf = NewConfig() // Global conf

type ZConfiguration struct {
	Stage                 Stage
	LogLevel              log.Level
	EthNode               string // the main eth node
	BackupNode            string // a backup node for special occasions
	RinkebyNode           string // Rinkeby network, used for tests
	Database              ZoroDB
	BlocksDelaySeconds    int
	PollingInterval       int
	BlocksInterval        int
	TwitterConsumerKey    string
	TwitterConsumerSecret string
	EtherscanKey          string
	Network               string
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
	STAGING
	PROD
)

func (s Stage) String() string {
	return [...]string{"TEST", "STAGING", "PROD"}[s]
}

// ENV variables
const (
	blocksDelay           = "BLOCKS_DELAY"
	dbHost                = "DB_HOST"
	dbName                = "DB_NAME"
	dbUsr                 = "DB_USR"
	dbPwd                 = "DB_PWD"
	ethNode               = "ETH_NODE"
	backupNode            = "BACKUP_NODE"
	rinkebyNode           = "RINKEBY_NODE"
	twitterConsumerKey    = "TWITTER_CONSUMER_KEY"
	twitterConsumerSecret = "TWITTER_CONSUMER_SECRET"
	network               = "NETWORK"
	pollingInterval       = "POLLING_INTERVAL"
	blocksInterval        = "BLOCKS_INTERVAL"
	etherscanKey          = "ETHERSCAN_KEY"
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
		zconfig.LogLevel = log.DebugLevel
	case "STAGING":
		zconfig.Stage = STAGING
		zconfig.LogLevel = log.DebugLevel
	case "PROD":
		zconfig.Stage = PROD
		zconfig.LogLevel = log.InfoLevel
	default:
		log.Fatal("local env STAGING must be TEST, STAGING or PROD")
	}

	zconfig.Database.TableTriggers = tableTriggers
	zconfig.Database.TableMatches = tableMatches
	zconfig.Database.TableOutcomes = tableOutcomes
	zconfig.Database.TableState = tableState
	zconfig.Database.TableActions = tableActions
	zconfig.Database.TableUsers = tableUsers
	zconfig.Database.Port = dbPort

	delay := os.Getenv(blocksDelay)
	intDelay, err := strconv.Atoi(delay)
	if delay == "" || err != nil {
		log.Fatalf("cannot use %s as block delay", delay)
	}
	zconfig.BlocksDelaySeconds = intDelay

	zconfig.Database.Host = os.Getenv(dbHost)
	if zconfig.Database.Host == "" {
		log.Fatal("no db host set in local env ", dbHost)
	}

	zconfig.Database.Name = os.Getenv(dbName)
	if zconfig.Database.Name == "" {
		log.Fatal("no db name set in local env ", dbName)
	}

	if zconfig.Stage == TEST && zconfig.Database.Name != "hal_test" {
		log.Fatalf("cannot use db %s with stage set to TEST\n", zconfig.Database.Name)
	}

	zconfig.Database.User = os.Getenv(dbUsr)
	if zconfig.Database.User == "" {
		log.Fatal("no db user set in local env ", dbUsr)
	}

	zconfig.Database.Password = os.Getenv(dbPwd)
	if zconfig.Database.Password == "" {
		log.Fatal("no db password set in local env ", dbPwd)
	}

	zconfig.Network = os.Getenv(network)
	if zconfig.Network == "" {
		log.Fatal("no network set in local env ", network)
	}

	zconfig.EthNode = os.Getenv(ethNode)
	if zconfig.EthNode == "" {
		log.Fatal("no eth node set in local env ", ethNode)
	}

	zconfig.BackupNode = os.Getenv(backupNode)
	if zconfig.BackupNode == "" {
		log.Fatal("no backup node set in local env ", backupNode)
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

	zconfig.EtherscanKey = os.Getenv(etherscanKey)
	if zconfig.EtherscanKey == "" {
		log.Fatal("no etherscan key set in local env ", etherscanKey)
	}

	interval := os.Getenv(pollingInterval)
	intervalSeconds, err := strconv.Atoi(interval)
	if interval == "" || err != nil {
		log.Fatalf("cannot use %s as polling interval", interval)
	}
	zconfig.PollingInterval = intervalSeconds

	blocksInterval := os.Getenv(blocksInterval)
	blocksIntervalSeconds, err := strconv.Atoi(blocksInterval)
	if blocksInterval == "" || err != nil {
		log.Fatalf("cannot use %s as blocks interval", blocksInterval)
	}
	zconfig.BlocksInterval = blocksIntervalSeconds

	return &zconfig
}

func (c ZConfiguration) IsNetworkETHMainnet() bool {
	return c.Network == "1_eth_mainnet"
}
