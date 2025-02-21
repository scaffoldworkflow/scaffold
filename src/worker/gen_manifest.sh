#! /usr/bin/env bash

kubectl -n scaffold delete cm scaffold-scripts || true
kubectl create configmap -n scaffold scaffold-scripts --from-file=.executor.sh --from-file=.header.sh --from-file=.header.py -o yaml > manifest.yaml
