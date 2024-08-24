package header

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Header struct {
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

func (h Header) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("header_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, h)
	return doc.String(), err
}
