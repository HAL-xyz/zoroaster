package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type ZConfiguration struct {
	Stage    string
	EthNode  string
	LogsPath string
	LogsFile string
	Database ZoroDB
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
const (
	dbUsr   = "DB_USR"
	dbPwd   = "DB_PWD"
	ethNode = "ETH_NODE"
)

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
