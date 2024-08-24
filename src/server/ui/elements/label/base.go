package label

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Label struct {
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

func (l Label) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("label_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, l)
	return doc.String(), err
}
