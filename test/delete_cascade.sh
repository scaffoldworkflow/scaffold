#! /usr/bin/env bash

source creds.sh

name="$1"

curl -X DELETE -H "Content-Type: application/json" -H "Authorization: X-Scaffold-API ${PRIMARY_KEY}" "http://localhost:2997/api/v1/cascade/${name}"
