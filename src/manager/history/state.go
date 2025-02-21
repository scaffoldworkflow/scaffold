package history

import (
	"encoding/json"
	"fmt"
	"scaffold/manager/constants"
	"strings"

	logger "github.com/jfcarter2358/go-logger"
)

type State struct {
	Step       string `json:"step" bson:"step" yaml:"step"`
	Status     string `json:"status" bson:"status" yaml:"status"`
	Started    string `json:"started" bson:"started" yaml:"started"`
	Finished   string `json:"finished" bson:"finished" yaml:"finished"`
	Output     string `json:"output" bson:"output" yaml:"output"`
	HTMLOutput string `json:"html_output" bson:"html_output" yaml:"html_output"`
	Idx        int    `json:"idx" bson:"idx" yaml:"idx"`
	RunID      string `json:"run_id" bson:"run_id" yaml:"run_id"`
}

func (s *State) ParseLogs(lines []string, previousCount int) map[string]interface{} {
	output := ""
	htmlOutput := ""
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
			htmlOutput += buildPre(line)
		}
	}
	s.Output = output
	s.HTMLOutput = htmlOutput

	return context
}

func buildPre(line string) string {
	line = replaceControlCodes(line)
	if strings.HasPrefix(line, "[FATAL]") {
		return fmt.Sprintf(`<pre class="ui-text-red"><code>%s</code></pre>`, line)
	}
	if strings.HasPrefix(line, "[SUCCESS]") {
		return fmt.Sprintf(`<pre class="ui-text-green"><code>%s</code></pre>`, line)
	}
	if strings.HasPrefix(line, "[ERROR]") {
		return fmt.Sprintf(`<pre class="ui-text-red"><code>%s</code></pre>`, line)
	}
	if strings.HasPrefix(line, "[WARN]") {
		return fmt.Sprintf(`<pre class="ui-text-yellow"><code>%s</code></pre>`, line)
	}
	if strings.HasPrefix(line, "[INFO]") {
		return fmt.Sprintf(`<pre class="ui-text-green"><code>%s</code></pre>`, line)
	}
	if strings.HasPrefix(line, "[DEBUG]") {
		return fmt.Sprintf(`<pre class="ui-text-cyan"><code>%s</code></pre>`, line)
	}
	if strings.HasPrefix(line, "[TRACE]") {
		return fmt.Sprintf(`<pre class="ui-text-purple"><code>%s</code></pre>`, line)
	}
	return fmt.Sprintf(`<pre><code>%s</code></pre>`, line)
}

func replaceControlCodes(line string) string {
	RED := "\033[0;31m"
	YELLOW := "\033[0;33m"
	GREEN := "\033[0;32m"
	CYAN := "\033[0;36m"
	BLUE := "\033[0;34m"
	NO_COLOR := "\033[0m"

	line = strings.ReplaceAll(line, RED, ``)
	line = strings.ReplaceAll(line, YELLOW, ``)
	line = strings.ReplaceAll(line, GREEN, ``)
	line = strings.ReplaceAll(line, CYAN, ``)
	line = strings.ReplaceAll(line, BLUE, ``)
	line = strings.ReplaceAll(line, NO_COLOR, ``)

	return line
}
