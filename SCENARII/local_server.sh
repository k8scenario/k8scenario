#!/bin/bash

cd $(dirname $0)/

# Defaults - Listen on localhost:9000 only:
PORT=9000
BIND="--bind 127.0.0.1"

echo BIND=$BIND PORT=$PORT

PYTHON=python3
$PYTHON -m http.server $PORT $BIND



