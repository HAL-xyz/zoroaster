package config

import (
	"github.com/HAL-xyz/zoroaster/rpc"
	"github.com/onrik/ethrpc"
)

// global variables across Zoroaster

var Zconf = NewConfig()

var cli = ethrpc.New(Zconf.EthNode)
var cliRinkeby = ethrpc.New(Zconf.RinkebyNode)

// used for templating only
var TemplateCli = rpc.New(cli, "templating client")

// used for tests only
var CliMain = rpc.New(cli, "test client")
var CliRinkeby = rpc.New(cliRinkeby, "rinkeby test client")
