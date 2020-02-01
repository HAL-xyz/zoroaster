package config

import "github.com/onrik/ethrpc"

// global variables across Zoroaster

var Zconf = Load()

var CliMain = ethrpc.New(Zconf.EthNode)
var CliRinkeby = ethrpc.New(Zconf.RinkebyNode)
