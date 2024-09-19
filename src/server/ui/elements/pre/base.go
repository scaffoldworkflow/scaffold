package pre

import (
	"bytes"
	_ "embed"
	"scaffold/server/ui"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Pre struct {
	ID         string
	Components []ui.Component
	Classes    string
	Style      string
	HXGet      string
	HXPost     string
	HXPut      string
	HXTarget   string
	HXSwap     string
	HXTrigger  string
}

var Template *template.Template

func (p Pre) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("div_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, p)
	return doc.String(), err
}
