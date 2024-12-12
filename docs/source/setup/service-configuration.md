# Service Configuration

Scaffold allows for a whole slew of configuration options that can be configured via environment variables in the deployment

All configuration ENV variables start with `SCAFFOLD_`

| Variable | Description | Default |
|---|---|---|
| SCAFFOLD_WS_PORT | <span style="color:red">DEPRECATED</span> Websocket port for task shell | 8080 |
| SCAFFOLD_LOG_LEVEL | Log level to use during operation. Valid levels are `NONE`, `DEBUG`, `ERROR`, `FATAL`, `INFO`, `SUCCESS`, `TRACE`, `WARN` | `INFO` |
| SCAFFOLD_LOG_FORMAT | Log format to output. Valid formats are `console`, `json` | `console` |
| SCAFFOLD_BASE_URL | Service base URL | `http://localhost:2997` |
| SCAFFOLD_PODMAN_OPTS | Options to pass to podman command in container tasks | `--security-opt label=disabled --network=host` |
| SCAFFOLD_ADMIN | Default admin user configuration | `{"username":"admin","password":"admin"}` |
| SCAFFOLD_DB_CONNECTION_STRING | Connection string for MongoDB | `mongodb://MyCoolMongoDBUsername:MyCoolMongoDBPassword@mongodb:27017/scaffold` |
| SCAFFOLD_NODE | Node information. For workers change the type to `worker` | `{"type":"manager","manager_host":"scaffold-manager","manager_port":2997,"manager_protocol":"http","join_key":"MyCoolJoinKey12345","primary_key":"MyCoolPrimaryKey12345"}` |
| SCAFFOLD_HEARTBEAT_INTERVAL | How frequently to check for worker health in seconds | `1000` |
| SCAFFOLD_HEARTBEAT_BACKOFF | How many retries before removing node | `10` |
| SCAFFOLD_RESET | Reset email configuration | `{"email":"","password":"","host":"smtp.gmail.com","port":587}` |
| SCAFFOLD_FILESTORE | Filestore configuration for use either with S3 or Artifactory. For Artifactory, change `host` and `port` to point to your Artifactory instance, `bucket` to the path in Artifactory to store the files, and `type` to `artifactory` | `{"access_key":"MyCoolMinIOAccessKey","secret_key":"MyCoolMinIOSecretKey","host":"minio","port":9000,"bucket":"scaffold","region":"default-region","protocol":"http","type":"s3"}` |
| SCAFFOLD_TLS_ENABLED | Should Scaffold run on https | `false` |
| SCAFFOLD_TLS_SKIP_VERIFY | Should Scaffold skip ssl verification in requests | `false` |
| SCAFFOLD_TLS_CERT_PATH | Path to mounted certificate for https | `/tmp/certs/cert.crt` |
| SCAFFOLD_TLS_KEY_PATH | Path to mounted key for https | `tmp/certs/cert.key` |
| SCAFFOLD_RABBITMQ_CONNECTION_STRING | Connection string for RabbitMQ | |
| SCAFFOLD_MANAGER_QUEUE_NAME | Name of the queue for messages to the manager | `scaffold_manager` |
| SCAFFOLD_WORKER_QUEUE_NAME | Name of the queue for messages to the worker | `scaffold_worker` |
| SCAFFOLD_KILL_QUEUE_NAME | Name of the queue to pass task runs to be killed | `scaffold_kill` |
| SCAFFOLD_PING_HEALTHY_THRESHOLD | How many pings until node is considered not healthy | `3` |
| SCAFFOLD_UNKNOWN_THRESHOLD | How many pings until node is considered unknown status | `6` |
| SCAFFOLD_PING_DOWN_THRESHOLD | How many pings until node is considered down | `9` |
| SCAFFOLD_CHECK_INTERVAL | How frequently to check task status in milliseconds | `2000` |
| SCAFFOLD_RESTART_PERIOD | How long in milliseconds before the service should restart itself. Set to `0` to disable automatic restarts | `86400` |
| SCAFFOLD_RUN_PRUNE_CRON | Crontab to prune run histories | `0 0 * * * *` |
| SCAFFOLD_RUN_PRUNE_DURATION | How long runs can stay around before being pruned in hours | `24` |
