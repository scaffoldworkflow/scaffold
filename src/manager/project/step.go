package project

type DependsOn struct {
	Success []string `json:"success" bson:"success" yaml:"success"`
	Error   []string `json:"error" bson:"error" yaml:"error"`
	Always  []string `json:"always" bson:"always" yaml:"always"`
}

type Step struct {
	Inputs       []string  `json:"inputs" bson:"inputs" yaml:"inputs"`
	Outputs      []string  `json:"outputs" bson:"outputs" yaml:"outputs"`
	Run          string    `json:"run" bson:"run" yaml:"run"`
	DependsOn    DependsOn `json:"depends_on" bson:"depends_on" yaml:"depends_on"`
	AutoExecute  bool      `json:"auto_execute" bson:"auto_execute" yaml:"auto_execute"`
	Requirements string    `json:"requirements" bson:"requirements" yaml:"requirements"`
	Language     string    `json:"language" bson:"language" yaml:"language"`
	LogLevel     string    `json:"log_level" bson:"log_level" yaml:"log_level"`
}
