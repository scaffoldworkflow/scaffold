package msg

import "scaffold/server/state"

type RunMsg struct {
	Task     string            `json:"task"`
	Workflow string            `json:"workflow"`
	Status   string            `json:"status"`
	RunID    string            `json:"run_id"`
	Context  map[string]string `json:"context"`
	State    state.State       `json:"state"`
}

type TriggerMsg struct {
	Task     string            `json:"task"`
	Workflow string            `json:"workflow"`
	Action   string            `json:"action"`
	Number   int               `json:"number"`
	Groups   []string          `json:"groups"`
	RunID    string            `json:"run_id"`
	Context  map[string]string `json:"context"`
}
