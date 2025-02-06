#! /usr/bin/env bash

curl -H "Authorization: Bearer MyCoolPrimaryKey12345" -X POST --data "$(cat test/manual/baz.yaml | yq -c '.')" http://localhost:2997/api/v1/project