package objects

type Cascade struct {
	Version string            `json:"version" bson:"version"`
	Name    string            `json:"name" bson:"name"`
	Inputs  []Input           `json:"inputs" bson:"inputs"`
	Tasks   []Task            `json:"tasks" bson:"tasks"`
	Created string            `json:"created" bson:"created"`
	Updated string            `json:"updated" bson:"updated"`
	Groups  []string          `json:"groups" bson:"groups"`
	Links   map[string]string `json:"links" bson:"links"`
}

type DataStore struct {
	Name    string            `json:"name" bson:"name"`
	Env     map[string]string `json:"env" bson:"env"`
	Files   []string          `json:"files" bson:"files"`
	Created string            `json:"created" bson:"created"`
	Updated string            `json:"updated" bson:"updated"`
}

type Input struct {
	Name        string `json:"name" bson:"name"`
	Cascade     string `json:"cascade" bson:"cascade"`
	Description string `json:"description" bson:"description"`
	Default     string `json:"default" bson:"default"`
	Type        string `json:"type" bson:"type"`
}

type State struct {
	Task     string                   `json:"task" bson:"task"`
	Cascade  string                   `json:"cascade" bson:"cascade"`
	Status   string                   `json:"status" bson:"status"`
	Started  string                   `json:"started" bson:"started"`
	Finished string                   `json:"finished" bson:"finished"`
	Output   string                   `json:"output" bson:"output"`
	Display  []map[string]interface{} `json:"display" bson:"display"`
	Number   int                      `json:"number" bson:"number"`
}

type TaskDependsOn struct {
	Success []string `json:"success" bson:"success"`
	Error   []string `json:"error" bson:"error"`
	Always  []string `json:"always" bson:"always"`
}

type TaskLoadStore struct {
	Env  []string `json:"env" bson:"env"`
	File []string `json:"file" bson:"file"`
}

type TaskCheck struct {
	Interval  int               `json:"interval" bson:"interval"`
	Image     string            `json:"image" bson:"image"`
	Run       string            `json:"run" bson:"run"`
	Store     TaskLoadStore     `json:"store" bson:"store"`
	Load      TaskLoadStore     `json:"load" bson:"load"`
	Env       map[string]string `json:"env" bson:"env"`
	Inputs    map[string]string `json:"inputs" bson:"inputs"`
	Updated   string            `json:"updated" bson:"updated"`
	RunNumber int               `json:"run_number" bson:"run_number"`
}

type Task struct {
	Name        string            `json:"name" bson:"name"`
	Cascade     string            `json:"cascade" bson:"cascade"`
	Verb        string            `json:"verb" bson:"verb"`
	DependsOn   TaskDependsOn     `json:"depends_on" bson:"depends_on"`
	Image       string            `json:"image" bson:"image"`
	Run         string            `json:"run" bson:"run"`
	Store       TaskLoadStore     `json:"store" bson:"store"`
	Load        TaskLoadStore     `json:"load" bson:"load"`
	Env         map[string]string `json:"env" bson:"env"`
	Inputs      map[string]string `json:"inputs" bson:"inputs"`
	Updated     string            `json:"updated" bson:"updated"`
	Check       TaskCheck         `json:"check" bson:"check"`
	RunNumber   int               `json:"run_number" bson:"run_number"`
	ShouldRM    bool              `json:"should_rm" bson:"should_rm"`
	AutoExecute bool              `json:"auto_execute" bson:"auto_execute"`
}
