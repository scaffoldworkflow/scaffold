package collapse

import (
	"bytes"
	_ "embed"
	"scaffold/server/ui"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Collapse struct {
	ID           string
	Components   []ui.Component
	Classes      string
	TitleClasses string
	TitleID      string
	TitleStyle   string
	Title        string
	Style        string
	HXGet        string
	HXPost       string
	HXPut        string
	HXTarget     string
	HXSwap       string
	HXTrigger    string
}

var Template *template.Template

func (c Collapse) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("collapse_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, c)
	return doc.String(), err
}
