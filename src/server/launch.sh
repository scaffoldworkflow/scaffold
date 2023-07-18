#! /usr/bin/env bash

RUN_DIR=$(dirname $0)

pushd "${RUN_DIR}"
./scaffold
popd
