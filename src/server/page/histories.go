package page

import (
	"net/http"
	"scaffold/server/history"
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

func HistoriesSearchEndpoint(ctx *gin.Context) {
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	histories, err := history.GetAllHistories()
	if err != nil {
		logger.Errorf("", "Cannot render histories page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	filtered := []history.History{}

	for _, h := range histories {
		if strings.Contains(strings.ToLower(h.RunID), strings.ToLower(query)) || strings.Contains(strings.ToLower(h.Workflow), strings.ToLower(query)) {
			filtered = append(filtered, *h)
		}
	}

	markdown := historiesBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func HistoriesTableEndpoint(ctx *gin.Context) {
	histories, err := history.GetAllHistories()
	if err != nil {
		logger.Errorf("", "Cannot render histories page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []history.History{}

	for _, h := range histories {
		filtered = append(filtered, *h)
	}

	markdown := historiesBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func HistoriesPageEndpoint(ctx *gin.Context) {
	markdown := historiesBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func historiesBuildPage(ctx *gin.Context) []byte {
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
										Title: "Runs",
										HRef:  "/ui/runs",
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: `<input id="search" class="w3-input w3-round search-bar theme-light" type="text"
                            name="search" placeholder="Search Runs"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%;width:98%" hx-get="/htmx/runs/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#histories-table-div" />`,
					},
					div.Div{
						ID:        "histories-table-div",
						HXTrigger: "load",
						HXGet:     "/htmx/runs/table",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render runs page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func historiesBuildTable(hs []history.History, ctx *gin.Context) []byte {
	sort.Slice(hs, func(i, j int) bool {
		return hs[i].RunID < hs[j].RunID
	})

	t := table.Table{
		ID: "histories_table",
		Headers: []header.Header{
			{
				Contents: "Name",
				Classes:  "text-lg",
			},
			{
				Contents: "Workflow",
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
				Contents: "Last Task",
				Classes:  "text-lg",
			},
			{
				Contents: "Last Status",
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
	ws := workflow.GetCacheAll()
	groups := []string{}
	for _, w := range ws {
		if w.Name == hs[0].Workflow {
			groups = w.Groups
		}
	}

	for _, h := range hs {
		isInGroup := false
		for _, ug := range u.Groups {
			if ug == "admin" {
				isInGroup = true
				break
			}
			for _, wg := range groups {
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
				break
			}
			if ur == "write" {
				break
			}
		}
		logger.Errorf("", "Current history check: %s", h.RunID)
		logger.Errorf("", "Current history states: %v", h.States)
		if len(h.States) > 0 {
			lastTask := h.States[len(h.States)-1].Task
			lastStatus := h.States[len(h.States)-1].Status
			r := []cell.Cell{
				{
					Contents: h.RunID,
				},
				{
					Contents: h.Workflow,
				},
				{
					Contents: h.Created,
				},
				{
					Contents: h.Updated,
				},
				{
					Contents: lastTask,
				},
				{
					Contents: lastStatus,
				},
				{
					Contents: `<a href="/ui/runs/` + h.RunID + `" class="table-link-link w3-right-align dark theme-text"
                    style="float:right;margin-right:16px;">
                    <i class="fa-solid fa-link"></i>
                </a>`,
				},
			}
			t.Rows = append(t.Rows, r)
		}
	}

	html, err := t.Render()
	if err != nil {
		logger.Errorf("", "Cannot render runs table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
