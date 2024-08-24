package br

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed template.html
var TemplateString string

type BR struct {
	ID string
}

var Template *template.Template

func (b BR) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("br_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, b)
	return doc.String(), err
}
