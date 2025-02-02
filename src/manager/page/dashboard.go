package page

import (
	"bytes"
	"html/template"
	"net/http"
	"scaffold/manager/project"

	_ "embed"

	logger "github.com/jfcarter2358/go-logger"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/gin-gonic/gin"
)

//go:embed dashboard-template.html
var dashboardTmplString string
var dashboardTmpl *template.Template

type DashboardProject struct {
	Name     string
	ColorHex string
}
type DashboardData struct {
	Projects []DashboardProject
}

func DashboardPage(ctx *gin.Context) {
	if dashboardTmpl == nil {
		var err error
		dashboardTmpl, err = template.New("k8s_job_template").Parse(dashboardTmplString)
		if err != nil {
			logger.Errorf("", "Cannot load k8s job template: %s", err.Error())
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	ps, err := project.GetProjects(bson.M{})
	if err != nil {
		logger.Errorf("", "Cannot not get projects: %s", err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	dashboardData := DashboardData{
		Projects: []DashboardProject{},
	}

	for _, p := range ps {
		dashboardData.Projects = append(dashboardData.Projects, DashboardProject{
			Name:     p.Name,
			ColorHex: "#009900",
		})
	}
	var doc bytes.Buffer
	if err := dashboardTmpl.Execute(&doc, dashboardData); err != nil {
		logger.Errorf("", "Cannot render k8s job template: %s", err.Error())
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	showPage(ctx, doc.String(), gin.H{})
}
