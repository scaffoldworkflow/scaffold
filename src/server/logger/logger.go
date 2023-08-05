package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"scaffold/server/constants"
	"time"

	"github.com/gin-gonic/gin"
)

var LogLevel = 0

var ConsoleLogFormatter = func(param gin.LogFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	timestamp := param.TimeStamp.Format("2006-01-02T15:04:05Z")
	statusCode := param.StatusCode
	latency := param.Latency
	clientIP := param.ClientIP
	method := param.Method
	path := param.Path
	errorMessage := param.ErrorMessage

	statusCodeColor := constants.COLOR_GREEN
	methodColor := constants.COLOR_MAGENTA

	level := constants.LOG_LEVEL_INFO
	if statusCode >= 400 {
		level = constants.LOG_LEVEL_ERROR
		statusCodeColor = constants.COLOR_RED
	}

	switch method {
	case "GET":
		methodColor = constants.METHOD_GET
	case "POST":
		methodColor = constants.METHOD_POST
	case "PUT":
		methodColor = constants.METHOD_PUT
	case "PATCH":
		methodColor = constants.METHOD_PATCH
	case "DELETE":
		methodColor = constants.METHOD_DELETE
	}

	return Logf(level,
		constants.LOG_FORMAT_CONSOLE,
		timestamp,
		"%s%3d%s | %13v | %15s | %s%-7s%s %#v | %s",
		statusCodeColor,
		statusCode,
		constants.COLOR_NONE,
		latency,
		clientIP,
		methodColor,
		method,
		constants.COLOR_NONE,
		path,
		errorMessage,
	)
}

func sliceIndex(list []string, val string) int {
	for i := 0; i < len(list); i++ {
		if list[i] == val {
			return i
		}
	}
	return -1
}

func SetLevel(level string) {
	levels := []string{constants.LOG_LEVEL_FATAL, constants.LOG_LEVEL_SUCCESS, constants.LOG_LEVEL_ERROR, constants.LOG_LEVEL_WARN, constants.LOG_LEVEL_INFO, constants.LOG_LEVEL_DEBUG, constants.LOG_LEVEL_TRACE}
	levelInt := sliceIndex(levels, level)
	if levelInt == -1 {
		Fatalf("Unknown log level %s", level)
	}
	LogLevel = levelInt
}

func Logf(level, formatter, timestamp, format string, args ...interface{}) string {
	if formatter == constants.LOG_FORMAT_JSON {
		log := map[string]interface{}{
			"level":     level,
			"timestamp": timestamp,
			"message":   fmt.Sprintf(format, args...),
		}
		logBytes, _ := json.Marshal(&log)
		return string(logBytes)
	}
	switch level {
	case constants.LOG_LEVEL_DEBUG:
		return Sdebugf(timestamp, format, args...)
	case constants.LOG_LEVEL_ERROR:
		return Serrorf(timestamp, format, args...)
	case constants.LOG_LEVEL_FATAL:
		return Sfatalf(timestamp, format, args...)
	case constants.LOG_LEVEL_INFO:
		return Sinfof(timestamp, format, args...)
	case constants.LOG_LEVEL_SUCCESS:
		return Ssuccessf(timestamp, format, args...)
	case constants.LOG_LEVEL_TRACE:
		return Stracef(timestamp, format, args...)
	case constants.LOG_LEVEL_WARN:
		return Swarnf(timestamp, format, args...)
	}
	return ""
}

func Debug(timestamp, message string) {
	if LogLevel >= constants.LOG_LEVEL_DEBUG_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[DEBUG  ]%s [%s] :: %s\n", constants.COLOR_CYAN, constants.COLOR_NONE, timestamp, message)
	}
}

func Debugf(timestamp, format string, args ...interface{}) {
	if LogLevel >= constants.LOG_LEVEL_DEBUG_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[DEBUG  ]%s [%s] :: %s\n", constants.COLOR_CYAN, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
}

func Sdebugf(timestamp, format string, args ...interface{}) string {
	if LogLevel >= constants.LOG_LEVEL_DEBUG_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		return fmt.Sprintf("%s[DEBUG  ]%s [%s] :: %s\n", constants.COLOR_CYAN, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
	return ""
}

func Error(timestamp, message string) {
	if LogLevel >= constants.LOG_LEVEL_ERROR_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[ERROR  ]%s [%s] :: %s\n", constants.COLOR_RED, constants.COLOR_NONE, timestamp, message)
	}
}

func Errorf(timestamp, format string, args ...interface{}) {
	if LogLevel >= constants.LOG_LEVEL_ERROR_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[ERROR  ]%s [%s] :: %s\n", constants.COLOR_RED, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
}

func Serrorf(timestamp, format string, args ...interface{}) string {
	if LogLevel >= constants.LOG_LEVEL_ERROR_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		return fmt.Sprintf("%s[ERROR  ]%s [%s] :: %s\n", constants.COLOR_RED, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
	return ""
}

func Fatal(timestamp, message string) {
	if LogLevel >= constants.LOG_LEVEL_FATAL_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[FATAL  ]%s [%s] :: %s\n", constants.COLOR_RED, constants.COLOR_NONE, timestamp, message)
		os.Exit(1)
	}
}

func Fatalf(timestamp, format string, args ...interface{}) {
	if LogLevel >= constants.LOG_LEVEL_FATAL_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[FATAL  ]%s [%s] :: %s\n", constants.COLOR_RED, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
		os.Exit(1)
	}
}

func Sfatalf(timestamp, format string, args ...interface{}) string {
	if LogLevel >= constants.LOG_LEVEL_FATAL_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		return fmt.Sprintf("%s[FATAL  ]%s [%s] :: %s\n", constants.COLOR_RED, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
	return ""
}

func Info(timestamp, message string) {
	if LogLevel >= constants.LOG_LEVEL_INFO_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[INFO   ]%s [%s] :: %s\n", constants.COLOR_GREEN, constants.COLOR_NONE, timestamp, message)
	}
}

func Infof(timestamp, format string, args ...interface{}) {
	if LogLevel >= constants.LOG_LEVEL_INFO_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[INFO   ]%s [%s] :: %s\n", constants.COLOR_GREEN, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
}

func Sinfof(timestamp, format string, args ...interface{}) string {
	if LogLevel >= constants.LOG_LEVEL_INFO_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		return fmt.Sprintf("%s[INFO   ]%s [%s] :: %s\n", constants.COLOR_GREEN, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
	return ""
}

func Success(timestamp, message string) {
	if LogLevel >= constants.LOG_LEVEL_SUCCESS_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[SUCCESS]%s [%s] :: %s\n", constants.COLOR_GREEN, constants.COLOR_NONE, timestamp, message)
	}
}

func Successf(timestamp, format string, args ...interface{}) {
	if LogLevel >= constants.LOG_LEVEL_SUCCESS_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[SUCCESS]%s [%s] :: %s\n", constants.COLOR_GREEN, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
}

func Ssuccessf(timestamp, format string, args ...interface{}) string {
	if LogLevel >= constants.LOG_LEVEL_SUCCESS_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		return fmt.Sprintf("%s[SUCCESS]%s [%s] :: %s\n", constants.COLOR_GREEN, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
	return ""
}

func Trace(timestamp, message string) {
	if LogLevel >= constants.LOG_LEVEL_TRACE_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[TRACE  ]%s [%s] :: %s\n", constants.COLOR_BLUE, constants.COLOR_NONE, timestamp, message)
	}
}

func Tracef(timestamp, format string, args ...interface{}) {
	if LogLevel >= constants.LOG_LEVEL_TRACE_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[TRACE  ]%s [%s] :: %s\n", constants.COLOR_BLUE, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
}

func Stracef(timestamp, format string, args ...interface{}) string {
	if LogLevel >= constants.LOG_LEVEL_TRACE_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		return fmt.Sprintf("%s[TRACE  ]%s [%s] :: %s\n", constants.COLOR_BLUE, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
	return ""
}

func Warn(timestamp, message string) {
	if LogLevel >= constants.LOG_LEVEL_WARN_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[WARN   ]%s [%s] :: %s\n", constants.COLOR_YELLOW, constants.COLOR_NONE, timestamp, message)
	}
}

func Warnf(timestamp, format string, args ...interface{}) {
	if LogLevel >= constants.LOG_LEVEL_WARN_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		fmt.Printf("%s[WARN   ]%s [%s] :: %s\n", constants.COLOR_YELLOW, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
}

func Swarnf(timestamp, format string, args ...interface{}) string {
	if LogLevel >= constants.LOG_LEVEL_WARN_NUM {
		if timestamp == "" {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}
		return fmt.Sprintf("%s[WARN   ]%s [%s] :: %s\n", constants.COLOR_YELLOW, constants.COLOR_NONE, timestamp, fmt.Sprintf(format, args...))
	}
	return ""
}
