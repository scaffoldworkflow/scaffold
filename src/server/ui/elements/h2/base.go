package h2

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed template.html
var TemplateString string

type H2 struct {
	ID        string
	Contents  string
	Classes   string
	Style     string
	HXGet     string
	HXPost    string
	HXPut     string
	HXTarget  string
	HXSwap    string
	HXTrigger string
}

var Template *template.Template

func (h H2) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("h2_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, h)
	return doc.String(), err
}
