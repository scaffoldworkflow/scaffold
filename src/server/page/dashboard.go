package page

import (
	"fmt"
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/state"
	"scaffold/server/ui"
	"scaffold/server/ui/breadcrumb"
	"scaffold/server/ui/elements/br"
	"scaffold/server/ui/elements/div"
	"scaffold/server/ui/elements/link"
	"scaffold/server/ui/page"
	"scaffold/server/ui/sidebar"
	"scaffold/server/ui/table"
	"scaffold/server/ui/table/cell"
	"scaffold/server/ui/table/header"
	"scaffold/server/ui/topbar"
	"scaffold/server/user"
	"scaffold/server/workflow"
	"sort"
	"strings"

	_ "embed"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func DashboardSearchEndpoint(ctx *gin.Context) {
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

	markdown := dashboardBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func DashboardTableEndpoint(ctx *gin.Context) {
	workflows := workflow.GetCacheAll()
	filtered := []workflow.Workflow{}

	for _, w := range workflows {
		filtered = append(filtered, w)
	}

	markdown := dashboardBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func DashboardPageEndpoint(ctx *gin.Context) {
	markdown := dashboardBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func dashboardBuildPage(ctx *gin.Context) []byte {
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
				Classes: "scaffold-green",
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
						Classes: "scaffold-green rounded-md",
						Components: []ui.Component{
							breadcrumb.Breadcrumb{
								Components: []ui.Component{
									link.Link{
										Title: "Dashboard",
										HRef:  "/ui/dashboard",
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: `<input id="search" class="w3-input w3-round search-bar theme-light" type="text"
                            name="search" placeholder="Search Workflows"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%;width:98%" hx-get="/htmx/dashboard/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#workflows-table-div" />`,
					},
					div.Div{
						ID:        "dashboard-table-div",
						HXTrigger: "load, every 1s",
						HXGet:     "/htmx/dashboard/table",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render dashboard page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func dashboardBuildTable(ws []workflow.Workflow, ctx *gin.Context) []byte {
	sort.Slice(ws, func(i, j int) bool {
		return ws[i].Name < ws[j].Name
	})

	t := table.Table{
		ID: "dashboard_table",
		Headers: []header.Header{
			{
				Contents: "Name",
				Classes:  "text-lg",
			},
			{
				Contents: "Task States",
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
		HeaderClasses: "rounded-md scaffold-green",
	}

	token, _ := ctx.Cookie("scaffold_token")
	u, _ := user.GetUserByLoginToken(token)

	for _, w := range ws {
		isInGroup := false
		for _, ug := range u.Groups {
			if ug == "admin" {
				isInGroup = true
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
		ss, err := state.GetAllStates()
		if err != nil {
			logger.Errorf("", "Cannot render dashboard table: %s", err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return []byte{}
		}
		runningCount := 0
		waitingCount := 0
		successCount := 0
		errorCount := 0
		notStartedCount := 0
		killedCount := 0
		// TODO: Figure out something better so this isn't running in O(n^2) time
		for _, s := range ss {
			if s.Workflow == w.Name {
				switch s.Status {
				case constants.STATE_STATUS_ERROR:
					errorCount += 1
				case constants.STATE_STATUS_KILLED:
					killedCount += 1
				case constants.STATE_STATUS_NOT_STARTED:
					notStartedCount += 1
				case constants.STATE_STATUS_RUNNING:
					runningCount += 1
				case constants.STATE_STATUS_SUCCESS:
					successCount += 1
				case constants.STATE_STATUS_WAITING:
					waitingCount += 1
				}
			}
		}
		total := runningCount + waitingCount + successCount + errorCount + notStartedCount + killedCount
		percentRunning := int(float64(runningCount) / float64(total) * 80.0)
		percentWaiting := int(float64(waitingCount) / float64(total) * 80.0)
		percentSuccess := int(float64(successCount) / float64(total) * 80.0)
		percentError := int(float64(errorCount) / float64(total) * 80.0)
		percentNotStarted := int(float64(notStartedCount) / float64(total) * 80.0)
		percentKilled := int(float64(killedCount) / float64(total) * 80.0)
		r := []cell.Cell{
			{
				Contents: w.Name,
			},
			{
				Contents: fmt.Sprintf(`
				<div class="scaffold-charcoal" style="width:%d%%;height:20px;display:inline-block;margin-top:8px;margin-left:-4px;"></div>
				<div class="scaffold-yellow" style="width:%d%%;height:20px;display:inline-block;margin-top:8px;margin-left:-4px;"></div>
				<div class="scaffold-blue" style="width:%d%%;height:20px;display:inline-block;margin-top:8px;margin-left:-4px;"></div>
				<div class="scaffold-green" style="width:%d%%;height:20px;display:inline-block;margin-top:8px;margin-left:-4px;"></div>
				<div class="scaffold-red" style="width:%d%%;height:20px;display:inline-block;margin-top:8px;margin-left:-4px;"></div>
				<div class="scaffold-orange" style="width:%d%%;height:20px;display:inline-block;margin-top:8px;margin-left:-4px;"></div>
				`, percentNotStarted, percentWaiting, percentRunning, percentSuccess, percentError, percentKilled),
			},
			{
				Contents: `<a href="/ui/workflows/` + w.Name + `" class="table-link-link w3-right-align dark theme-text"
                    style="float:right;margin-right:16px;">
                    <i class="fa-solid fa-link"></i>
                </a>`,
			},
		}
		t.Rows = append(t.Rows, r)
	}

	html, err := t.Render()
	if err != nil {
		logger.Errorf("", "Cannot render dashboard table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
