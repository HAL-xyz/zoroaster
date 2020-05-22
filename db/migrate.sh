#!/usr/bin/env bash

if [[ -z "$1" ]]
  then
    echo "Usage: $0 <up|down>"
    echo "To run N migration(s) up or down, pass \"up N\" as input"
    echo "You need to export the right local variables. See source"
    exit
fi

POSTGRESQL_URL="postgres://${DB_USR}:${DB_PWD}@${DB_URI}/${DB_NAME}?sslmode=disable"

read -p "You are going to migrate the db ${DB_NAME}; continue? [y\n] " CHOICE
if [[ "$CHOICE" != "y" ]]; then
    echo "Aborting"
    exit
fi

migrate -database ${POSTGRESQL_URL} -path migrations $1
