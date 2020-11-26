#!/usr/bin/env bash

export KO_DATA_PATH=${KO_DATA_PATH:-"kodata/"}
export PREFIX=${PREFIX:-"/surfdash/"}
export PORT=${PORT:-8080}
go run main.go
