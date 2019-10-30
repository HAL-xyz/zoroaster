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
./zoroaster
```

You can build for a specific architecture, e.g.
```
env GOOS=linux GOARCH=amd64 go build -o zoroaster
```

## Tests

You can run the tests for a specifc package with `go test` from within that package, or you can run all tests and generate a `cover.html` file using the `run_tests.sh` script.

## License
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
