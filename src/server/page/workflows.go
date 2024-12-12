package page

import (
	"net/http"
	"scaffold/server/user"
	"scaffold/server/workflow"
	"sort"
	"strings"

	"github.com/jfcarter2358/ui"
	"github.com/jfcarter2358/ui/breadcrumb"
	"github.com/jfcarter2358/ui/elements/br"
	"github.com/jfcarter2358/ui/elements/div"
	"github.com/jfcarter2358/ui/elements/link"
	"github.com/jfcarter2358/ui/page"
	"github.com/jfcarter2358/ui/sidebar"
	"github.com/jfcarter2358/ui/table"
	"github.com/jfcarter2358/ui/table/cell"
	"github.com/jfcarter2358/ui/table/header"
	"github.com/jfcarter2358/ui/topbar"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func WorkflowsSearchEndpoint(ctx *gin.Context) {
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	workflows := workflow.GetCacheAll()

	filtered := []workflow.Workflow{}

	for name, w := range workflows {
		if strings.Contains(strings.ToLower(name), strings.ToLower(query)) {
			filtered = append(filtered, w)
		}
	}

	markdown := workflowsBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func WorkflowsTableEndpoint(ctx *gin.Context) {
	workflows := workflow.GetCacheAll()
	filtered := []workflow.Workflow{}

	for _, w := range workflows {
		filtered = append(filtered, w)
	}

	markdown := workflowsBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func WorkflowsPageEndpoint(ctx *gin.Context) {
	markdown := workflowsBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func workflowsBuildPage(ctx *gin.Context) []byte {
	p := page.Page{
		ID:             "page",
		SidebarEnabled: true,
		Sidebar: sidebar.Sidebar{
			ID:      "sidebar",
			Classes: "theme-light",
			Components: []ui.Component{
				link.Link{
					Title: "Dashboard",
					HRef:  "/ui/dashboard",
				},
				link.Link{
					Title: "Runs",
					HRef:  "/ui/runs",
				},
				link.Link{
					Title: "Users",
					HRef:  "/ui/users",
				},
				link.Link{
					Title: "Workflows",
					HRef:  "/ui/workflows",
				},
			},
		},
		Components: []ui.Component{
			topbar.Topbar{
				Title:   "Scaffold",
				Classes: "ui-green",
				Buttons: []ui.Component{
					link.Link{
						Title:   "Logout",
						HRef:    "/auth/logout",
						Style:   "passing:12px;",
						Classes: "theme-dark rounded-md",
					},
				},
				MenuClasses: "theme-light",
			},
			div.Div{
				Classes: "theme-light rounded-md",
				Components: []ui.Component{
					div.Div{
						Classes: "ui-green rounded-md",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Workflows",
										HRef:  "/ui/workflows",
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: `<input id="search" class="w3-input w3-round search-bar theme-light" type="text"
                            name="search" placeholder="Search Workflows"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%;width:98%" hx-get="/htmx/workflows/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#workflows-table-div" />`,
					},
					div.Div{
						ID:        "workflows-table-div",
						HXTrigger: "load",
						HXGet:     "/htmx/workflows/table",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render workflows page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func workflowsBuildTable(ws []workflow.Workflow, ctx *gin.Context) []byte {
	sort.Slice(ws, func(i, j int) bool {
		return ws[i].Name < ws[j].Name
	})

	t := table.Table{
		ID: "workflows_table",
		Headers: []header.Header{
			{
				Contents: "Name",
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
				Contents: "Version",
				Classes:  "text-lg",
			},
			{
				Contents: "",
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
		HeaderClasses: "rounded-md ui-green",
	}

	token, _ := ctx.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	for _, w := range ws {
		isInGroup := false
		isAdmin := false
		for _, ug := range u.Groups {
			if ug == "admin" {
				isInGroup = true
				isAdmin = true
				break
			}
			for _, wg := range w.Groups {
				if ug == wg {
					isInGroup = true
					break
				}
			}
			if isInGroup {
				break
			}
		}
		for _, ur := range u.Roles {
			if ur == "admin" {
				isAdmin = true
				break
			}
			if ur == "write" {
				break
			}
		}
		r := []cell.Cell{
			{
				Contents: w.Name,
			},
			{
				Contents: w.Created,
			},
			{
				Contents: w.Updated,
			},
			{
				Contents: w.Version,
			},
			{
				Contents: `<a href="/ui/workflows/` + w.Name + `" class="table-link-link w3-right-align dark theme-text"
                    style="float:right;margin-right:16px;">
                    <i class="fa-solid fa-link"></i>
                </a>`,
			},
			{
				Contents: "",
			},
		}
		if isAdmin {
			r[len(r)-1] = cell.Cell{
				Contents: `<div class="table-link-link w3-right-align dark theme-text" style="float:right;margin-right:16px;">
                    <i class="fa-solid fa-trash" style="cursor:pointer;" onclick="deleteWorkflow('` + w.Name + `')"></i>
                </div>`,
			}
		}
		t.Rows = append(t.Rows, r)
	}

	html, err := t.Render()
	if err != nil {
		logger.Errorf("", "Cannot render workflows table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
