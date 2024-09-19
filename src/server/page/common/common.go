package common

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/manager"
	"scaffold/server/user"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed status.html
var statusHTML string
var statusTemplate *template.Template

//go:embed sidebar.html
var sidebarHTML string
var sidebarTemplate *template.Template

//go:embed error.html
var errorHTML string
var errorTemplate *template.Template

//go:embed success.html
var successHTML string
var successTemplate *template.Template

//go:embed 401.html
var code401HTML string
var code401Template *template.Template

//go:embed 403.html
var code403HTML string
var code403Template *template.Template

//go:embed 404.html
var code404HTML string
var code404Template *template.Template

//go:embed 500.html
var code500HTML string
var code500Template *template.Template

//go:embed header.html
var headerHTML string
var headerTemplate *template.Template

func Init() error {
	var err error
	statusTemplate, err = template.New("common__status").Parse(statusHTML)
	if err != nil {
		return err
	}
	sidebarTemplate, err = template.New("common__sidebar").Parse(sidebarHTML)
	if err != nil {
		return err
	}
	errorTemplate, err = template.New("common__error").Parse(errorHTML)
	if err != nil {
		return err
	}
	successTemplate, err = template.New("common__success").Parse(successHTML)
	if err != nil {
		return err
	}
	headerTemplate, err = template.New("common__header").Parse(headerHTML)
	if err != nil {
		return err
	}
	code401Template, err = template.New("common__401").Parse(code401HTML)
	if err != nil {
		return err
	}
	code403Template, err = template.New("common__403").Parse(code403HTML)
	if err != nil {
		return err
	}
	code404Template, err = template.New("common__404").Parse(code404HTML)
	if err != nil {
		return err
	}
	code500Template, err = template.New("common__500").Parse(code500HTML)
	if err != nil {
		return err
	}
	return nil
}

func StatusEndpoint(ctx *gin.Context) {
	isHealthy, nodes := manager.GetStatus()
	var markdown bytes.Buffer
	type Data struct {
		IsHealthy bool
		Nodes     []manager.UINode
	}
	statusTemplate.Execute(&markdown, Data{IsHealthy: isHealthy, Nodes: nodes})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}

func SidebarEndpoint(ctx *gin.Context) {
	var markdown bytes.Buffer
	type Data struct {
		Version string
	}
	sidebarTemplate.Execute(&markdown, Data{Version: constants.VERSION})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}

func ErrorEndpoint(ctx *gin.Context) {
	var markdown bytes.Buffer
	type Data struct{}
	errorTemplate.Execute(&markdown, Data{})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}

func SuccessEndpoint(ctx *gin.Context) {
	var markdown bytes.Buffer
	type Data struct{}
	successTemplate.Execute(&markdown, Data{})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}

func Code401Endpoint(ctx *gin.Context) {
	var markdown bytes.Buffer
	type Data struct{}
	code401Template.Execute(&markdown, Data{})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}

func Code403Endpoint(ctx *gin.Context) {
	var markdown bytes.Buffer
	type Data struct{}
	code403Template.Execute(&markdown, Data{})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}

func Code404Endpoint(ctx *gin.Context) {
	var markdown bytes.Buffer
	type Data struct{}
	code404Template.Execute(&markdown, Data{})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}

func Code500Endpoint(ctx *gin.Context) {
	var markdown bytes.Buffer
	type Data struct{}
	code500Template.Execute(&markdown, Data{})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}

func HeaderEndpoint(ctx *gin.Context) {
	link, ok := ctx.GetQuery("link")
	if !ok {
		ctx.Redirect(http.StatusTemporaryRedirect, "/ui/404")
	}
	caser := cases.Title(language.English)
	title := caser.String(link)

	var markdown bytes.Buffer
	type Data struct {
		Title      string
		Link       string
		GivenName  string
		FamilyName string
	}
	token, _ := ctx.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	familyName := ""
	givenName := ""
	if u != nil {
		familyName = u.FamilyName
		givenName = u.GivenName
	}
	headerTemplate.Execute(&markdown, Data{
		Title:      title,
		Link:       link,
		FamilyName: familyName,
		GivenName:  givenName,
	})
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown.Bytes())
}
