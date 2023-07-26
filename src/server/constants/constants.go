package constants

const VERSION = "0.1.0"

const STATE_STATUS_ERROR = "error"
const STATE_STATUS_SUCCESS = "success"
const STATE_STATUS_RUNNING = "running"
const STATE_STATUS_WAITING = "waiting"
const STATE_STATUS_NOT_STARTED = "not_started"

const MONGODB_CASCADE_COLLECTION_NAME = "cascade"
const MONGODB_DATASTORE_COLLECTION_NAME = "datastore"
const MONGODB_STATE_COLLECTION_NAME = "state"
const MONGODB_USER_COLLECTION_NAME = "user"
const MONGODB_TASK_COLLECTION_NAME = "task"
const MONGODB_INPUT_COLLECTION_NAME = "input"

const NODE_TYPE_WORKER = "worker"
const NODE_TYPE_MANAGER = "manager"
