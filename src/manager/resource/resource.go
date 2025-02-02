package resource

import (
	_ "embed"
)

//go:embed file.py
var FileScript string

//go:embed github_repo.py
var GitHubRepoScript string

type Resource struct {
	Name string            `json:"name" bson:"name" yaml:"name"`
	Kind string            `json:"kind" bson:"kind" yaml:"kind"`
	Args map[string]string `json:"args" bson:"args" yaml:"args"`
}

type ResourceType struct {
	Name         string `json:"name" bson:"name" yaml:"name"`
	Source       string `json:"source" bson:"source" yaml:"source"`
	Script       string `json:"script" bson:"script"`
	Language     string `json:"language" bson:"language" yaml:"language"`
	Requirements string `json:"requirements" bson:"requirements" yaml:"requirements"`
}
