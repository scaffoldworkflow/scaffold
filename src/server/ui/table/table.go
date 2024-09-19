package table

import (
	"bytes"
	_ "embed"
	"text/template"

	"scaffold/server/ui/table/cell"
	"scaffold/server/ui/table/header"
)

//go:embed template.html
var TemplateString string

type Table struct {
	ID            string
	Headers       []header.Header
	Rows          [][]cell.Cell
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

func (t Table) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("table_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, t)
	return doc.String(), err
}
