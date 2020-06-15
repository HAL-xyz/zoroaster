package config

import "github.com/onrik/ethrpc"

// global variables across Zoroaster

var Zconf = NewConfig()

var CliMain = ethrpc.New(Zconf.EthNode)
var CliRinkeby = ethrpc.New(Zconf.RinkebyNode)
