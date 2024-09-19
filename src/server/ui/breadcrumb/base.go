package breadcrumb

import (
	"bytes"
	_ "embed"
	"scaffold/server/ui"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Breadcrumb struct {
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

func (b Breadcrumb) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("breadcrumb_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, b)
	return doc.String(), err
}
