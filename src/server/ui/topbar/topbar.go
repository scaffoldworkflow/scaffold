package topbar

import (
	"bytes"
	_ "embed"
	"scaffold/server/ui"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Topbar struct {
	ID          string
	Components  []ui.Component
	Classes     string
	Style       string
	Links       []ui.Component
	Title       string
	Buttons     []ui.Component
	MenuClasses string
}

var Template *template.Template

func (t Topbar) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("topbar_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, t)
	return doc.String(), err
}
