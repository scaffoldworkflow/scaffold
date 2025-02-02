package constants

const VERSION = "0.4.2.alpha1"

const STATE_STATUS_ERROR = "error"
const STATE_STATUS_SUCCESS = "success"
const STATE_STATUS_RUNNING = "running"
const STATE_STATUS_WAITING = "waiting"
const STATE_STATUS_NOT_STARTED = "not_started"
const STATE_STATUS_KILLED = "killed"

const MONGODB_USER_COLLECTION_NAME = "user"
const MONGODB_HISTORY_COLLECTION_NAME = "history"
const MONGODB_RUNBOOK_COLLECTION_NAME = "runbook"
const MONGODB_ALERT_COLLECTION_NAME = "alert"
const MONGODB_PROJECT_COLLECTION_NAME = "project"
const MONGODB_ENVIRONMENT_COLLECTION_NAME = "environment"
const MONGODB_SERVICE_COLLECTION_NAME = "service"

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

const NODE_HEALTHY = "healthy"
const NODE_DEGRADED = "degraded"
const NODE_UNHEALTHY = "unhealthy"
const NODE_UNKNOWN = "unknown"
const NODE_WARN = "warn"
const NODE_ERROR = "error"
const NODE_DEPLOYING = "deploying"
const NODE_NOT_DEPLOYED = "not-deployed"
const NODE_KILLED = "killed"
const NODE_WAITING = "waiting"
const NODE_SUCCESS = "success"
const NODE_RUNNING = "running"

const MONITOR_STATUS_RUNNING = "running"
const MONITOR_STATUS_STOPPED = "stopped"
const MONITOR_STATUS_ALERT = "alert"

const ALERT_STATUS_RUNNING = "running"
const ALERT_STATUS_KILLED = "killed"
const ALERT_STATUS_ERROR = "error"

const ALERT_STATUS_ENABLED = "enabled"
const ALERT_STATUS_PAUSED = "paused"

const MONITOR_KIND_K8S = "k8s"
const MONITOR_KIND_FILE = "file"
const MONITOR_KIND_CUSTOM = "custom"

var UI_HEALTH_ICONS = map[string]string{
	NODE_HEALTHY:   "fa-circle-check",
	NODE_DEGRADED:  "fa-circle-exclamation",
	NODE_UNHEALTHY: "fa-circle-xmark",
	NODE_UNKNOWN:   "fa-circle-question",
}

var UI_HEALTH_COLORS = map[string]string{
	NODE_HEALTHY:   "ui-text-green",
	NODE_DEGRADED:  "ui-text-yellow",
	NODE_UNHEALTHY: "ui-text-red",
	NODE_UNKNOWN:   "ui-text-charcoal",
}

var UI_HEALTH_TEXT = map[string]string{
	NODE_HEALTHY:   "Up",
	NODE_DEGRADED:  "Degraded",
	NODE_UNHEALTHY: "Down",
	NODE_UNKNOWN:   "Unknown",
}

var UI_ICONS = map[string]string{
	NODE_HEALTHY:      "fa-circle-check",
	NODE_WARN:         "fa-circle-exclamation",
	NODE_ERROR:        "fa-circle-xmark",
	NODE_DEPLOYING:    "fa-spinner fa-pulse",
	NODE_UNKNOWN:      "fa-circle-question",
	NODE_NOT_DEPLOYED: "fa-circle",
}

var UI_COLORS = map[string]string{
	NODE_HEALTHY:      "green",
	NODE_WARN:         "yellow",
	NODE_ERROR:        "red",
	NODE_DEPLOYING:    "grey",
	NODE_UNKNOWN:      "orange",
	NODE_KILLED:       "orange",
	NODE_WAITING:      "yellow",
	NODE_NOT_DEPLOYED: "grey",
	NODE_RUNNING:      "blue",
	NODE_SUCCESS:      "green",
}

const MONITOR_SLEEP = 1

const KERNEL_RESPONSE_SUCCESS = "kernel::response::success"
const KERNEL_RESPONSE_ERROR = "kernel::response::error"
const KERNEL_RESPONSE_STOPPED = "kernel::response::stopped"
const KERNEL_RESPONSE_KILLED = "kernel::response::killed"

const KERNEL_COMMAND_KILL = "kernel::command::kill"

const KERNEL_SCRIPT_START = "kernel::script::start"
const KERNEL_SCRIPT_END = "kernel::script::end"

const KERNEL_OUTPUT_PREFIX = "kernel::output::"

const KERNEL_LANG_PYTHON = "kernel::lang::python"
const KERNEL_LANG_BASH = "kernel::lang::bash"

const KERNEL_RESOURCE_INPUT = "kernel::resource::input"
const KERNEL_RESOURCE_OUTPUT = "kernel::resource::output"
const KERNEL_RESOURCE_CONTEXT = "kernel::resource::context"
