package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type ZConfiguration struct {
	Stage      string
	EthNode    string
	LogsPath   string
	LogsFile   string
	TriggersDB TriggersDB
}

type TriggersDB struct {
	TableTriggers string
	TableMatches  string
	TableOutcomes string
	TableStats    string
	TableActions  string
	Host          string
	User          string
	Name          string
	Port          int
	Password      string
}

func Load(dirpath string) *ZConfiguration {

	var zconfig ZConfiguration
	var configFile string
	stage := os.Getenv("STAGE")
	switch stage {
	case "DEV":
		configFile = fmt.Sprintf("%s/config-dev.json", dirpath)
		zconfig.Stage = "DEV"
	case "PROD":
		configFile = fmt.Sprintf("%s/config-prod.json", dirpath)
		zconfig.Stage = "PROD"
	default:
		log.Fatal("local env STAGE must be DEV or PROD")
	}

	var err error
	f, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("cannot open %s: %s", configFile, err)
	}
	err = json.Unmarshal(f, &zconfig)
	if err != nil {
		log.Fatalf("cannot load %s: %s", configFile, err)
	}

	zconfig.LogsFile = fmt.Sprintf("%s/%s.log", zconfig.LogsPath, stage)

	const dbPwd = "DB_PWD"
	zconfig.TriggersDB.Password = os.Getenv(dbPwd)
	if zconfig.TriggersDB.Password == "" {
		log.Fatal("no db password set in local env ", dbPwd)
	}

	return &zconfig
}
