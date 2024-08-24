package link

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Link struct {
	ID        string
	Title     string
	HRef      string
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

func (l Link) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("link_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, l)
	return doc.String(), err
}
