package history

import (
	"encoding/json"
	"scaffold/manager/constants"
	"strings"

	logger "github.com/jfcarter2358/go-logger"
)

type State struct {
	Step     string `json:"step" bson:"step" yaml:"step"`
	Status   string `json:"status" bson:"status" yaml:"status"`
	Started  string `json:"started" bson:"started" yaml:"started"`
	Finished string `json:"finished" bson:"finished" yaml:"finished"`
	Output   string `json:"output" bson:"output" yaml:"output"`
	Idx      int    `json:"idx" bson:"idx" yaml:"idx"`
	RunID    string `json:"run_id" bson:"run_id" yaml:"run_id"`
}

func (s *State) ParseLogs(lines []string, previousCount int) map[string]interface{} {
	output := ""
	context := make(map[string]interface{})
	for _, line := range lines {
		if strings.HasPrefix(line, constants.KERNEL_RESOURCE_CONTEXT) {
			contextJSON := line[len(constants.KERNEL_RESOURCE_CONTEXT)+2:]
			if err := json.Unmarshal([]byte(contextJSON), &context); err != nil {
				logger.Errorf("", "Cannot unmarshal context JSON '%s': %s", contextJSON, err.Error())
				continue
			}
		} else {
			output += line
		}
	}
	s.Output = output
	return context
}
