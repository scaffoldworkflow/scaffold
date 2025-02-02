package page

import (
	"bytes"
	"html/template"
	"net/http"
	"scaffold/manager/constants"
	"scaffold/manager/project"
	"scaffold/manager/user"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/bson"

	_ "embed"
)

//go:embed base-template.html
var baseTmplString string
var baseTmpl *template.Template

func RedirectIndexPage(c *gin.Context) {
	c.Redirect(301, "/ui/dashboard")
}

func showPage(c *gin.Context, contents string, header gin.H) {
	token, _ := c.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	familyName := ""
	givenName := ""
	if u != nil {
		familyName = u.FamilyName
		givenName = u.GivenName
	}

	ps, err := project.GetProjects(bson.M{})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	teams := map[string][]string{}
	if ps != nil {
		for _, p := range ps {
			if _, ok := teams[p.Team]; !ok {
				teams[p.Team] = []string{p.Name}
				continue
			}
			teams[p.Team] = append(teams[p.Team], p.Name)
		}
	}

	header["family_name"] = familyName
	header["given_name"] = givenName
	header["version"] = constants.VERSION
	header["teams"] = teams
	header["contents"] = contents

	render(c, header, baseTmplString)
}

func render(c *gin.Context, data gin.H, templateName string) {
	switch c.Request.Header.Get("Accept") {
	case "application/json":
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		c.XML(http.StatusOK, data["payload"])
	default:
		if baseTmpl == nil {
			var err error
			baseTmpl, err = template.New("k8s_job_template").Parse(baseTmplString)
			if err != nil {
				logger.Errorf("", "Cannot load page template: %s", err.Error())
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		}

		var doc bytes.Buffer
		if err := dashboardTmpl.Execute(&doc, data); err != nil {
			logger.Errorf("", "Cannot render page template: %s", err.Error())
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", doc.Bytes())
	}
}
