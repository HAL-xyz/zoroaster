# Zoroaster
A (work in progress) Golang daemon to execute Trigger->Action workflows based on the Ethereum blockchain.

## Install
Install all the dependencies:
```
go get -d -t ./...
```
then build with
```
go build -o zoroaster
```

You can build for a specific architecture, e.g.
```
env GOOS=linux GOARCH=amd64 go build -o zoroaster
```

## Run


1. Copy the appropriate configuration file(s) under `config`
into your local `/etc/` directory.
2. You need to export the following local variables:
   * `STAGE` - can be TEST, DEV or PROD
   * `DB_USR` 
   * `DB_PWD`
   * `ETH_NODE` - a valid Ethereum node
   * `TEST_NODE` - can be the same as the main `ETH_NODE`
   * `RINKEBY_NODE` - valid Ethereum Rinkeby node
   
Then you need to create a suitable database schema.
Fill in the `db/migrate_up.sh` script, then run it like this:
```
./migrate_up.sh <up|down>
```

finally, run Zoroaster:

```
./zoroaster
```

## Tests

You can run the tests for a specifc package with `go test` from within that package, or you can run all tests and generate a `cover.html` file using the `run_tests.sh` script.
Note that you can only run tests when the local `STAGE` variable is set to `TEST`.

## License
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
