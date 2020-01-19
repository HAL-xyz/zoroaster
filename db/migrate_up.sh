#!/usr/bin/env bash

DB_USR=""
DB_PWD=""
DB_URI=""
DB_NAME=""

if [[ -z "$1" ]]
  then
    echo "Usage: $0 <up|down>"
    exit
fi

POSTGRESQL_URL="postgres://${DB_USR}:${DB_PWD}@${DB_URI}/${DB_NAME}?sslmode=disable"

echo "Using Postgres URL: ${POSTGRESQL_URL}"

migrate -database ${POSTGRESQL_URL} -path migrations $1
