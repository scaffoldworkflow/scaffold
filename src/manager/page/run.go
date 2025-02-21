package page

import (
	"fmt"
	"net/http"
	"scaffold/manager/constants"
	"scaffold/manager/history"
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

func RunSearchEndpoint(ctx *gin.Context) {
	releaseID := ctx.Param("release_id")
	runID := ctx.Param("run_id")
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	logger.Tracef("", "Getting histories for %s", releaseID)
	hs, err := history.GetHistories(bson.M{"release_id": releaseID, "run_id": runID})
	if err != nil {
		logger.Errorf("", "Cannot get history at %s: %s", releaseID, err)
		ctx.Error(err)
		return
	}

	filtered := []history.History{}

	for _, h := range hs {
		if strings.Contains(strings.ToLower(h.Service), strings.ToLower(query)) || strings.Contains(strings.ToLower(h.Created), strings.ToLower(query)) || strings.Contains(strings.ToLower(h.Updated), strings.ToLower(query)) {
			filtered = append(filtered, *h)
		}
	}

	markdown := runBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func RunTableEndpoint(ctx *gin.Context) {
	releaseID := ctx.Param("release_id")
	runID := ctx.Param("run_id")
	hs, err := history.GetHistories(bson.M{"release_id": releaseID, "run_id": runID})
	if err != nil {
		logger.Errorf("", "Cannot render run page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []history.History{}

	for _, h := range hs {
		filtered = append(filtered, *h)
	}

	logger.Tracef("", "Got %d histories ", len(hs))

	markdown := runBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func RunPageEndpoint(ctx *gin.Context) {
	markdown := runBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func runBuildPage(ctx *gin.Context) []byte {
	releaseID := ctx.Param("release_id")
	runID := ctx.Param("run_id")
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
										Title: "Releases",
										HRef:  "/ui/releases",
									},
									link.Link{
										Title: releaseID,
										HRef:  fmt.Sprintf("/ui/releases/%s", releaseID),
									},
									link.Link{
										Title: "Runs",
										HRef:  fmt.Sprintf("/ui/releases/%s/%s", releaseID, runID),
									},
									link.Link{
										Title: releaseID,
										HRef:  fmt.Sprintf("/ui/releases/%s/%s", releaseID, runID),
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input id="search" class="w3-input search-bar theme-light" type="text"
                            name="search" placeholder="Search Services"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%%;width:98%%" hx-get="/htmx/run/search/%s/%s"
                            hx-trigger="keyup changed delay:250ms" hx-target="#run-table-div" />`, releaseID, runID),
					},
					div.Div{
						ID:        "run-table-div",
						HXTrigger: "load",
						HXGet:     fmt.Sprintf("/htmx/run/table/%s/%s", releaseID, runID),
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := pp.Render()
	if err != nil {
		logger.Errorf("", "Cannot render run page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func runBuildTable(hs []history.History, ctx *gin.Context) []byte {
	releaseID := ctx.Param("release_id")
	runID := ctx.Param("run_id")

	sort.Slice(hs, func(i, j int) bool {
		return hs[i].Updated > hs[j].Updated
	})

	sort.Slice(hs, func(i, j int) bool {
		return hs[i].Service < hs[j].Service
	})

	t := table.Table{
		ID: "releases_table",
		Headers: []header.Header{
			{
				Contents: "Service",
				Classes:  "text-lg",
			},
			{
				Contents: "Status",
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

	for _, h := range hs {
		status := constants.STATE_STATUS_NOT_STARTED
		if len(h.States) > 0 {
			status = h.States[len(h.States)-1].Status
		}
		// r := m[p]
		r := []cell.Cell{
			{
				Contents: h.Service,
			},
			{
				Contents: status,
			},
			{
				Contents: h.Created,
			},
			{
				Contents: h.Updated,
			},
			{
				Contents: fmt.Sprintf(`<a href="/ui/releases/%s/%s/%s" class="table-link-link w3-right-align dark theme-text"
				style="float:right;margin-right:16px;">
				<i class="fa-solid fa-link"></i>
			</a>`, releaseID, runID, h.Service),
			},
		}
		t.Rows = append(t.Rows, r)
	}

	d := div.Div{
		ID:         "run-table-div",
		Components: []ui.Component{t},
	}

	html, err := d.Render()
	if err != nil {
		logger.Errorf("", "Cannot render run table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
