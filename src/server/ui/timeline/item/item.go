package item

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Item struct {
	ID          string
	BoxContents string
	BoxClasses  string
	BoxStyle    string
	IconClasses string
	IconStyle   string
	LineColor   string
	HXGet       string
	HXPost      string
	HXPut       string
	HXTarget    string
	HXSwap      string
	HXTrigger   string
	IsFirst     bool
	IsLast      bool
}

var Template *template.Template

func (i Item) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("item_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, i)
	return doc.String(), err
}
