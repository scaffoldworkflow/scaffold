package constants

const VERSION = "0.3.0"

const STATE_STATUS_ERROR = "error"
const STATE_STATUS_SUCCESS = "success"
const STATE_STATUS_RUNNING = "running"
const STATE_STATUS_WAITING = "waiting"
const STATE_STATUS_NOT_STARTED = "not_started"
const STATE_STATUS_KILLED = "killed"

const MONGODB_CASCADE_COLLECTION_NAME = "cascade"
const MONGODB_DATASTORE_COLLECTION_NAME = "datastore"
const MONGODB_STATE_COLLECTION_NAME = "state"
const MONGODB_USER_COLLECTION_NAME = "user"
const MONGODB_TASK_COLLECTION_NAME = "task"
const MONGODB_INPUT_COLLECTION_NAME = "input"
const MONGODB_WEBHOOK_COLLECTION_NAME = "webhook"

const NODE_TYPE_WORKER = "worker"
const NODE_TYPE_MANAGER = "manager"

const COLOR_RED = "\033[0;31m"
const COLOR_YELLOW = "\033[0;33m"
const COLOR_GREEN = "\033[0;32m"
const COLOR_CYAN = "\033[0;36m"
const COLOR_BLUE = "\033[0;34m"
const COLOR_NONE = "\033[0m"
const COLOR_MAGENTA = "\033[0;35m"

const METHOD_GET = COLOR_YELLOW
const METHOD_POST = COLOR_GREEN
const METHOD_PUT = COLOR_BLUE
const METHOD_PATCH = COLOR_CYAN
const METHOD_DELETE = COLOR_RED

const STATUS_TRIGGER_ALWAYS = "always"
const STATUS_TRIGGER_SUCCESS = "success"
const STATUS_TRIGGER_ERROR = "error"

const ACTION_TRIGGER = "trigger"
const ACTION_KILL = "kill"

const FILESTORE_TYPE_S3 = "s3"
const FILESTORE_TYPE_ARTIFACTORY = "artifactory"

const TASK_KIND_LOCAL = "local"
const TASK_KIND_CONTAINER = "container"
