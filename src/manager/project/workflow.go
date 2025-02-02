package project

import (
	"scaffold/manager/resource"
	"time"

	_ "embed"
)

//go:embed default_job_template.yaml
var defaultJobTemplate string

type Workflow struct {
	Version       string                           `json:"version" bson:"version" yaml:"version"`
	Resources     map[string]resource.Resource     `json:"resources" bson:"resources" yaml:"resources"`
	ResourceTypes map[string]resource.ResourceType `json:"resource_types" bson:"resource_types" yaml:"resource_types"`
	Requirements  string                           `json:"requirements" bson:"requirements" yaml:"requirements"`
	Steps         map[string]Step                  `json:"steps" bson:"steps" yaml:"steps"`
	Created       string                           `json:"created" bson:"created" yaml:"created"`
	Updated       string                           `json:"updated" bson:"updated" yaml:"updated"`
	JobTemplate   string                           `json:"job_template" bson:"job_template" yaml:"job_template"`
}

func (w *Workflow) Load() error {
	currentTime := time.Now().UTC()

	// Create empty structures if not provided to avoid nils
	if w.Resources == nil {
		w.Resources = make(map[string]resource.Resource)
	}
	if w.ResourceTypes == nil {
		w.ResourceTypes = make(map[string]resource.ResourceType)
	}
	if w.Steps == nil {
		w.Steps = make(map[string]Step)
	}

	// Fill empty fields with defaults
	if w.JobTemplate == "" {
		w.JobTemplate = defaultJobTemplate
	}

	w.Created = currentTime.Format("2006-01-02T15:04:05Z")
	w.Updated = currentTime.Format("2006-01-02T15:04:05Z")
	return nil
}

func (w *Workflow) Update() error {

	currentTime := time.Now().UTC()

	// Create empty structures if not provided to avoid nils
	if w.Resources == nil {
		w.Resources = make(map[string]resource.Resource)
	}
	if w.ResourceTypes == nil {
		w.ResourceTypes = make(map[string]resource.ResourceType)
	}
	if w.Steps == nil {
		w.Steps = make(map[string]Step)
	}

	// Fill empty fields with defaults
	if w.JobTemplate == "" {
		w.JobTemplate = defaultJobTemplate
	}

	w.Updated = currentTime.Format("2006-01-02T15:04:05Z")

	return nil
}
