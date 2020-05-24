#!/usr/bin/env bash

if [[ -z "$1" ]]
  then
    echo "Usage: $0 <name of your new migration>"
    exit
fi

migrate create -ext sql -dir migrations/ -seq $1
