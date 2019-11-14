package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type ZConfiguration struct {
	Stage      Stage
	ConfigFile string
	EthNode    string
	LogsPath   string
	LogsFile   string
	Database   ZoroDB
}

type ZoroDB struct {
	TableTriggers string
	TableMatches  string
	TableOutcomes string
	TableState    string
	TableActions  string
	Host          string
	User          string
	Name          string
	Port          int
	Password      string
}

// ENV variables

type Stage int

const (
	TEST Stage = iota
	DEV
	PROD
)

func (s Stage) String() string {
	return [...]string{"TEST", "DEV", "PROD"}[s]
}

const (
	dbUsr   = "DB_USR"
	dbPwd   = "DB_PWD"
	ethNode = "ETH_NODE"
)

func readStage(dirPath string) ZConfiguration {
	zconf := ZConfiguration{}
	stage := os.Getenv("STAGE")
	switch stage {
	case "TEST":
		zconf.ConfigFile = fmt.Sprintf("%s/config-test.json", dirPath)
		zconf.Stage = TEST
	case "DEV":
		zconf.ConfigFile = fmt.Sprintf("%s/config-dev.json", dirPath)
		zconf.Stage = DEV
	case "PROD":
		zconf.ConfigFile = fmt.Sprintf("%s/config-prod.json", dirPath)
		zconf.Stage = PROD
	default:
		log.Fatal("local env STAGE must be TEST, DEV or PROD")
	}
	return zconf
}

func Load(dirPath string) *ZConfiguration {

	zconfig := readStage(dirPath)

	f, err := ioutil.ReadFile(zconfig.ConfigFile)
	if err != nil {
		log.Fatalf("cannot open %s: %s", zconfig.ConfigFile, err)
	}
	err = json.Unmarshal(f, &zconfig)
	if err != nil {
		log.Fatalf("cannot load %s: %s", zconfig.ConfigFile, err)
	}

	zconfig.LogsFile = fmt.Sprintf("%s/%s.log", zconfig.LogsPath, zconfig.Stage)

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

	return &zconfig
}
