package ui

import (
	_ "embed"
)

var PageTemplate string

type Component interface {
	Render() (string, error)
}

type Raw struct {
	HTMLString string
}

func (r Raw) Render() (string, error) {
	return r.HTMLString, nil
}
