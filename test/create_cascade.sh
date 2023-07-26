#! /usr/bin/env bash

source creds.sh

name="$1"

cat "${name}.yaml" | yq -c . > "/tmp/${name}.json"

curl -X POST -H "Content-Type: application/json" -H "Authorization: X-Scaffold-API ${PRIMARY_KEY}" -d "@/tmp/${name}.json" http://localhost:2997/api/v1/cascade

rm "/tmp/${name}.json"

