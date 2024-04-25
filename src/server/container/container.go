package container

import (
	"fmt"
	"os"
	"os/exec"
	"scaffold/server/run"
	"strings"
	"time"

	logger "github.com/jfcarter2358/go-logger"
)

var LastRun []string
var LastImage []string
var LastGroups [][]string
var CompletedRuns map[string]run.Run
var CurrentRun run.Run
var CurrentName string
var MaxAllowed = 10

func InitContainers() {
	LastRun = make([]string, 0)
	LastImage = make([]string, 0)
	LastGroups = make([][]string, 0)
	CurrentName = ""
}

func PruneContainers() {
	for {
		for len(LastRun) > MaxAllowed {
			toDestroy := ""
			toDestroy, LastRun = LastRun[0], LastRun[1:]
			LastImage = LastImage[1:]
			LastGroups = LastGroups[1:]

			parts := strings.Split(toDestroy, ".")

			if len(parts) == 2 {
				containerName := fmt.Sprintf("%s-%s", parts[0], parts[1])
				if containerName != CurrentName {
					logs, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("podman rm %s", containerName)).CombinedOutput()
					if err != nil {
						logger.Errorf("", "Remove error string: %s", err.Error())
					}
					logger.Debugf("", "Remove output: %s", logs)
					runDir := fmt.Sprintf("/tmp/run/%s/%s/%s", parts[0], parts[1], parts[2])
					if err := os.RemoveAll(runDir); err != nil {
						logger.Errorf("", "Delete error string: %s", err.Error())
					}
				}
			} else {
				logger.Debugf("", "Got %v when expected length 2", parts)
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
}
