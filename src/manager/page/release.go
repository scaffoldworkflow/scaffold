package page

// import (
// 	"fmt"
// 	"net/http"
// 	"scaffold/manager/constants"
// 	"scaffold/manager/history"
// 	"sort"

// 	"github.com/jfcarter2358/ui"
// 	"github.com/jfcarter2358/ui/breadcrumb"
// 	"github.com/jfcarter2358/ui/elements/br"
// 	"github.com/jfcarter2358/ui/elements/div"
// 	"github.com/jfcarter2358/ui/elements/link"
// 	"github.com/jfcarter2358/ui/page"
// 	"github.com/jfcarter2358/ui/table"
// 	"github.com/jfcarter2358/ui/table/cell"
// 	"github.com/jfcarter2358/ui/table/header"
// 	"github.com/jfcarter2358/ui/topbar"
// 	"go.mongodb.org/mongo-driver/bson"

// 	_ "embed"

// 	"github.com/gin-gonic/gin"
// 	logger "github.com/jfcarter2358/go-logger"
// )

// // func ReleaseSearchEndpoint(ctx *gin.Context) {
// // 	searchTerm, ok := ctx.GetQuery("search")
// // 	if !ok {
// // 		ctx.Status(http.StatusBadRequest)
// // 		return
// // 	}
// // 	query := strings.TrimSpace(searchTerm)

// // 	releases, err := release.GetReleases(bson.M{})
// // 	if err != nil {
// // 		logger.Errorf("", "Cannot render releases page: %s", err.Error())
// // 		ctx.AbortWithStatus(http.StatusInternalServerError)
// // 		return
// // 	}

// // 	filtered := []release.Release{}

// // 	for _, r := range releases {
// // 		if strings.Contains(strings.ToLower(r.Name), strings.ToLower(query)) || strings.Contains(strings.ToLower(r.Status), strings.ToLower(query)) || strings.Contains(strings.ToLower(r.Project), strings.ToLower(query)) || strings.Contains(strings.ToLower(r.Status), strings.ToLower(query)) {
// // 			filtered = append(filtered, *r)
// // 		}
// // 	}

// // 	markdown := releasesBuildTable(filtered, ctx)

// // 	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
// // }

// func ReleaseTableEndpoint(ctx *gin.Context) {
// 	pName := ctx.Param("project")
// 	tName := ctx.Param("team")
// 	eName := ctx.Param("environment")
// 	rID := ctx.Param("run_id")
// 	histories, err := history.GetHistories(bson.M{"team": tName, "project": pName, "environment": eName, "run_id": rID})
// 	if err != nil {
// 		logger.Errorf("", "Cannot render release page: %s", err.Error())
// 		ctx.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}
// 	filtered := []history.History{}

// 	for _, h := range histories {
// 		filtered = append(filtered, *h)
// 	}

// 	markdown := releaseBuildTable(filtered, tName, pName, eName, rID, ctx)

// 	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
// }

// func ReleasePageEndpoint(ctx *gin.Context) {
// 	markdown := releaseBuildPage(ctx)
// 	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
// }

// func releaseBuildPage(ctx *gin.Context) []byte {
// 	pName := ctx.Param("project")
// 	tName := ctx.Param("team")
// 	eName := ctx.Param("environment")
// 	rID := ctx.Param("run_id")
// 	pp := page.Page{
// 		ID:             "page",
// 		SidebarEnabled: true,
// 		Sidebar:        Sidebar,
// 		Components: []ui.Component{
// 			topbar.Topbar{
// 				Title:   "Scaffold",
// 				Classes: "ui-green",
// 				Buttons: []ui.Component{
// 					link.Link{
// 						Title:   "Logout",
// 						HRef:    "/auth/logout",
// 						Style:   "padding:12px;",
// 						Classes: "theme-dark",
// 					},
// 				},
// 				MenuClasses: "theme-light",
// 			},
// 			div.Div{
// 				Classes: "theme-light",
// 				Components: []ui.Component{
// 					div.Div{
// 						Classes: "ui-green",
// 						Components: []ui.Component{
// 							breadcrumb.Breadcrumb{
// 								Components: []ui.Component{
// 									link.Link{
// 										Title: "Releases",
// 										HRef:  "/ui/releases",
// 									},
// 									link.Link{
// 										Title: tName,
// 										HRef:  fmt.Sprintf("/ui/releases/%s/%s", tName, pName),
// 									},
// 									link.Link{
// 										Title: pName,
// 										HRef:  fmt.Sprintf("/ui/releases/%s/%s", tName, pName),
// 									},
// 									link.Link{
// 										Title: eName,
// 										HRef:  fmt.Sprintf("/ui/releases/%s/%s/%s/%s", tName, pName, eName, rID),
// 									},
// 									link.Link{
// 										Title: rID,
// 										HRef:  fmt.Sprintf("/ui/releases/%s/%s/%s/%s", tName, pName, eName, rID),
// 									},
// 								},
// 								Style: "margin-left:16px;",
// 							},
// 						},
// 					},
// 					ui.Raw{
// 						HTMLString: fmt.Sprintf(`<input id="search" class="w3-input search-bar theme-light" type="text"
//                             name="search" placeholder="Search Releases"
//                             style="margin-top:8px;margin-bottom:8px;margin-left:1%%;width:98%%" hx-get="/htmx/releases/%s/%s/%s/%s/search"
//                             hx-trigger="keyup changed delay:250ms" hx-target="#releases-table-div" />`, tName, pName, eName, rID),
// 					},
// 					div.Div{
// 						ID:        "releases-table-div",
// 						HXTrigger: "load",
// 						HXGet:     fmt.Sprintf("/htmx/releases/%s/%s/%s/%s/table", tName, pName, eName, rID),
// 					},
// 				},
// 				Style: "margin:64px;",
// 			},
// 			br.BR{},
// 		},
// 	}
// 	html, err := pp.Render()
// 	if err != nil {
// 		logger.Errorf("", "Cannot render releases page: %s", err.Error())
// 		ctx.AbortWithStatus(http.StatusInternalServerError)
// 		return []byte{}
// 	}
// 	return []byte(html)
// }

// func releaseBuildTable(hs []history.History, tName, pName, eName, rID string, ctx *gin.Context) []byte {
// 	sort.Slice(hs, func(i, j int) bool {
// 		return hs[i].Service < hs[j].Service
// 	})

// 	t := table.Table{
// 		ID: "releases_table",
// 		Headers: []header.Header{
// 			{
// 				Contents: "Name",
// 				Classes:  "text-lg",
// 			},
// 			{
// 				Contents: "Status",
// 				Classes:  "text-lg",
// 			},
// 			{
// 				Contents: "",
// 				Classes:  "text-lg",
// 			},
// 		},
// 		Rows:          make([][]cell.Cell, 0),
// 		Classes:       "theme-light",
// 		Style:         "width:100%;",
// 		HeaderClasses: "ui-green",
// 	}

// 	// token, _ := ctx.Cookie("scaffold_token")
// 	// u, _ := user.GetUserByLoginToken(token)

// 	for _, h := range hs {
// 		logger.Tracef("", "Going through history %v", h)
// 		status := constants.STATE_STATUS_NOT_STARTED
// 		if len(h.States) > 0 {
// 			status = h.States[len(h.States)-1].Status
// 		}
// 		r := []cell.Cell{
// 			{
// 				Contents: h.Service,
// 			},
// 			{
// 				Contents: status,
// 			},
// 			{
// 				Contents: fmt.Sprintf(`<a href="/ui/releases/%s/%s/%s/%s/%s" class="table-link-link w3-right-align dark theme-text"
// 							style="float:right;margin-right:16px;">
// 							<i class="fa-solid fa-link"></i>
// 						</a>`, tName, pName, eName, rID, h.Service),
// 			},
// 		}
// 		t.Rows = append(t.Rows, r)
// 	}

// 	html, err := t.Render()
// 	if err != nil {
// 		logger.Errorf("", "Cannot render releases table: %s", err.Error())
// 		ctx.AbortWithStatus(http.StatusInternalServerError)
// 		return []byte{}
// 	}
// 	return []byte(html)
// }

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

func ReleaseSearchEndpoint(ctx *gin.Context) {
	releaseID := ctx.Param("release_id")
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	logger.Tracef("", "Getting histories for %s", releaseID)
	hs, err := history.GetHistories(bson.M{"release_id": releaseID})
	if err != nil {
		logger.Errorf("", "Cannot get history at %s: %s", releaseID, err)
		ctx.Error(err)
		return
	}

	filtered := []history.History{}

	for _, h := range hs {
		status := constants.STATE_STATUS_NOT_STARTED
		if len(h.States) > 0 {
			status = h.States[len(h.States)-1].Status
		}
		if strings.Contains(strings.ToLower(h.RunID), strings.ToLower(query)) || strings.Contains(strings.ToLower(h.Created), strings.ToLower(query)) || strings.Contains(strings.ToLower(h.Updated), strings.ToLower(query)) || strings.Contains(strings.ToLower(status), strings.ToLower(query)) {
			filtered = append(filtered, *h)
		}
	}

	markdown := releaseBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func ReleaseTableEndpoint(ctx *gin.Context) {
	releaseID := ctx.Param("release_id")
	hs, err := history.GetHistories(bson.M{"release_id": releaseID})
	if err != nil {
		logger.Errorf("", "Cannot render release page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []history.History{}

	for _, h := range hs {
		filtered = append(filtered, *h)
	}

	logger.Tracef("", "Got %d histories ", len(hs))

	markdown := releaseBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func ReleasePageEndpoint(ctx *gin.Context) {
	markdown := releaseBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func releaseBuildPage(ctx *gin.Context) []byte {
	releaseID := ctx.Param("release_id")
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
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input id="search" class="w3-input search-bar theme-light" type="text"
                            name="search" placeholder="Search Runs"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%%;width:98%%" hx-get="/htmx/release/search/%s"
                            hx-trigger="keyup changed delay:250ms" hx-target="#release-table-div" />`, releaseID),
					},
					div.Div{
						ID:        "release-table-div",
						HXTrigger: "load",
						HXGet:     fmt.Sprintf("/htmx/release/table/%s", releaseID),
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := pp.Render()
	if err != nil {
		logger.Errorf("", "Cannot render release page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func releaseBuildTable(hs []history.History, ctx *gin.Context) []byte {
	releaseID := ctx.Param("release_id")

	sort.Slice(hs, func(i, j int) bool {
		return hs[i].Updated > hs[j].Updated
	})

	t := table.Table{
		ID: "releases_table",
		Headers: []header.Header{
			{
				Contents: "ID",
				Classes:  "text-lg",
			},
			{
				Contents: "Team",
				Classes:  "text-lg",
			},
			{
				Contents: "Project",
				Classes:  "text-lg",
			},
			{
				Contents: "Status",
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
				Contents: h.RunID,
			},
			{
				Contents: h.Team,
			},
			{
				Contents: h.Project,
			},
			{
				Contents: status,
			},
			{
				Contents: fmt.Sprintf(`<a href="/ui/releases/%s/%s" class="table-link-link w3-right-align dark theme-text"
				style="float:right;margin-right:16px;">
				<i class="fa-solid fa-link"></i>
			</a>`, releaseID, h.RunID),
			},
		}
		t.Rows = append(t.Rows, r)
	}
	// }

	d := div.Div{
		ID:         "releases-table-div",
		Components: []ui.Component{t},
	}

	html, err := d.Render()
	if err != nil {
		logger.Errorf("", "Cannot render releases table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
