package msg

type RunMsg struct {
	Task    string            `json:"task"`
	Cascade string            `json:"cascade"`
	Status  string            `json:"status"`
	Context map[string]string `json:"context"`
}

type TriggerMsg struct {
	Task    string            `json:"task"`
	Cascade string            `json:"cascade"`
	Action  string            `json:"action"`
	Number  int               `json:"number"`
	Groups  []string          `json:"groups"`
	Context map[string]string `json:"context"`
}
