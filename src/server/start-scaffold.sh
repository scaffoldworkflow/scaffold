#! /usr/bin/env bash

SLEEP_AMOUNT="${SCAFFOLD_SLEEP:-"0"}"
echo "Sleeping for ${SLEEP_AMOUNT} seconds"
sleep "${SLEEP_AMOUNT}"

# /bin/start-docker.sh

RUN_DIR=$(dirname $0)

pushd "${RUN_DIR}"
./scaffold
popd
