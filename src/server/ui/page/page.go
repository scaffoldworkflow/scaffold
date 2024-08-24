package page

import (
	"bytes"
	_ "embed"
	"scaffold/server/ui"
	"scaffold/server/ui/sidebar"
	"text/template"
)

//go:embed template.html
var TemplateString string

type Page struct {
	ID             string
	Components     []ui.Component
	Sidebar        sidebar.Sidebar
	SidebarEnabled bool
}

var Template *template.Template

func (p Page) Render() (string, error) {
	if Template == nil {
		var err error
		Template, err = template.New("page_html_template").Parse(TemplateString)
		if err != nil {
			return "", err
		}
	}
	var doc bytes.Buffer
	err := Template.Execute(&doc, p)
	return doc.String(), err
}
