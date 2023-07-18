package objects

type Job struct {
	ID      string   `json:"id" yaml:"id"`
	Name    string   `json:"name" yaml:"name"`
	Created string   `json:"created" yaml:"created"`
	Updated string   `json:"updated" yaml:"updated"`
	Steps   []string `json:"steps"`
}

type Pipeline struct {
	ID        string   `json:"id" yaml:"id"`
	Name      string   `json:"name" yaml:"name"`
	Created   string   `json:"created" yaml:"created"`
	Updated   string   `json:"updated" yaml:"updated"`
	Resources []string `json:"resources"`
	Jobs      []string `json:"jobs"`
}

type Resource struct {
	ID      string `json:"id" yaml:"id"`
	Name    string `json:"name" yaml:"name"`
	Created string `json:"created" yaml:"created"`
	Updated string `json:"updated" yaml:"updated"`
}

type Step struct {
	ID        string            `json:"id" yaml:"id"`
	Name      string            `json:"name" yaml:"name"`
	Created   string            `json:"created" yaml:"created"`
	Updated   string            `json:"updated" yaml:"updated"`
	Type      string            `json:"type" yaml:"type"`
	Resources []string          `json:"resources" yaml:"resources"`
	Contents  string            `json:"contents" yaml:"contents"`
	Image     string            `json:"image" yaml:"image"`
	Inputs    map[string]string `json:"inputs" yaml:"inputs"`
	Outputs   map[string]string `json:"outputs" yaml:"outputs"`
}
