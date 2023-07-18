#! /usr/bin/env bash

cat foobar.yaml | yq -c . > /tmp/foobar.json

curl -X POST -H "Content-Type: application/json" -d @/tmp/cascade.json http://localhost:2997/api/v1/cascade

rm /tmp/foobar.json

