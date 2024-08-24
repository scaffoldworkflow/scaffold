package timeline

import (
	"bytes"
	_ "embed"
	"scaffold/server/ui/timeline/item"

	"text/template"
)

//go:embed template.html
var TemplateString string

type Timeline struct {
	ID            string
	Items         []item.Item
	Classes       string
	Style         string
	HeaderClasses string
	HeaderStyle   string
	HXGet         string
	HXPost        string
	HXPut         string
	HXTarget      string
	HXSwap        string
	HXTrigger     string
}

var Template *template.Template

func (t Timeline) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("timeline_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, t)
	return doc.String(), err
}
