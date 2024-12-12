# Quickstart

## Deploying Docker Compose Scaffold

Scaffold works with a few different resources, all contained within the top-level docker-compose file in the [GitHub repository](https://github.com/scaffoldworkflow/scaffold):

- RabbitMQ
- MongoDB
- MinIO

To get the compose file, run the following:

```sh
curl -O https://raw.githubusercontent.com/scaffoldworkflow/scaffold/refs/heads/main/docker-compose.yaml
```

After doing so, run

```sh
docker-compose up -d
```

to start up Scaffold. Afterwards, access [http://localhost:2997](http://localhost:2997) and you should see the login page show up

## Installing the Scaffold CLI

Not that you have Scaffold running on your system, you need the CLI to interact with it. 

To install the latest version of the CLi, see the [GitHub Releases](https://github.com/scaffoldworkflow/scaffold/releases)

Once you've downloaded the appropriate binary for your machine, rename it to `scaffold` and copy it to your PATH. Afterwards, restart your shell and you should be able to run

```sh
> scaffold version local
Scaffold CLI Version: 0.4.1
```

```sh
> scaffold version remote
Scaffold Remote Version: 0.4.1
```

Now configure access to Scaffold:

```sh
scaffold configure --username admin --password admin
```

and you should be ready for your first workflow
