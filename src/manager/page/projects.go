package page

import (
	"fmt"
	"net/http"
	"scaffold/manager/project"
	"sort"
	"strings"

	"github.com/jfcarter2358/ui"
	"github.com/jfcarter2358/ui/breadcrumb"
	"github.com/jfcarter2358/ui/elements/br"
	"github.com/jfcarter2358/ui/elements/div"
	"github.com/jfcarter2358/ui/elements/link"
	"github.com/jfcarter2358/ui/page"
	"github.com/jfcarter2358/ui/table"
	"github.com/jfcarter2358/ui/table/cell"
	"github.com/jfcarter2358/ui/table/header"
	"github.com/jfcarter2358/ui/topbar"
	"go.mongodb.org/mongo-driver/bson"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func ProjectsSearchEndpoint(ctx *gin.Context) {
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	projects, err := project.GetProjects(bson.M{})
	if err != nil {
		logger.Errorf("", "Cannot render releases page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	filtered := []project.Project{}

	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(query)) || strings.Contains(strings.ToLower(p.Team), strings.ToLower(query)) {
			filtered = append(filtered, *p)
		}
	}

	markdown := projectsBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func ProjectsTableEndpoint(ctx *gin.Context) {
	projects, err := project.GetProjects(bson.M{})
	if err != nil {
		logger.Errorf("", "Cannot render projects page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []project.Project{}

	for _, p := range projects {
		filtered = append(filtered, *p)
	}

	markdown := projectsBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func ProjectsPageEndpoint(ctx *gin.Context) {
	markdown := projectsBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func projectsBuildPage(ctx *gin.Context) []byte {
	pp := page.Page{
		ID:             "page",
		SidebarEnabled: true,
		Sidebar:        Sidebar,
		Components: []ui.Component{
			topbar.Topbar{
				Title:   "Scaffold",
				Classes: "ui-green",
				Buttons: []ui.Component{
					link.Link{
						Title:   "Logout",
						HRef:    "/auth/logout",
						Style:   "padding:12px;",
						Classes: "theme-dark",
					},
				},
				MenuClasses: "theme-light",
			},
			div.Div{
				Classes: "theme-light",
				Components: []ui.Component{
					div.Div{
						Classes: "ui-green",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Projects",
										HRef:  "/ui/projects",
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: `<input id="search" class="w3-input search-bar theme-light" type="text"
                            name="search" placeholder="Search Projects"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%%;width:98%%" hx-get="/htmx/projects/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#projects-table-div" />`,
					},
					div.Div{
						ID:        "projects-table-div",
						HXTrigger: "load",
						HXGet:     "/htmx/projects/table",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := pp.Render()
	if err != nil {
		logger.Errorf("", "Cannot render projects page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func projectsBuildTable(ps []project.Project, ctx *gin.Context) []byte {
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Name < ps[j].Name
	})

	t := table.Table{
		ID: "projects_table",
		Headers: []header.Header{
			{
				Contents: "Name",
				Classes:  "text-lg",
			},
			{
				Contents: "Team",
				Classes:  "text-lg",
			},
			{
				Contents: "Created",
				Classes:  "text-lg",
			},
			{
				Contents: "Updated",
				Classes:  "text-lg",
			},
			{
				Contents: "",
				Classes:  "text-lg",
			},
		},
		Rows:          make([][]cell.Cell, 0),
		Classes:       "theme-light",
		Style:         "width:100%;",
		HeaderClasses: "ui-green",
	}

	// token, _ := ctx.Cookie("scaffold_token")
	// u, _ := user.GetUserByLoginToken(token)

	for _, p := range ps {
		logger.Tracef("", "Going through project %v", p)

		r := []cell.Cell{
			{
				Contents: p.Name,
			},
			{
				Contents: p.Team,
			},
			{
				Contents: p.Created,
			},
			{
				Contents: p.Updated,
			},
			{
				Contents: fmt.Sprintf(`<a href="/ui/projects/%s" class="table-link-link w3-right-align dark theme-text"
							style="float:right;margin-right:16px;">
							<i class="fa-solid fa-link"></i>
						</a>`, p.Name),
			},
		}
		t.Rows = append(t.Rows, r)
	}

	html, err := t.Render()
	if err != nil {
		logger.Errorf("", "Cannot render projects table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
