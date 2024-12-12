# Installation

Currently, Scaffold installation is only supported via docker-compose (kubernetes deployments are working but manifests have not yet been added to the repo)

To install via docker-compose, download the example file (edit it to your liking of course) and then deploy

```sh
curl -O https://raw.githubusercontent.com/scaffoldworkflow/scaffold/refs/heads/main/docker-compose.yaml
docker-compose up -d
```
