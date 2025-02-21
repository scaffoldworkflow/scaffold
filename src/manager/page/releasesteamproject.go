package page

import (
	"fmt"
	"net/http"
	"scaffold/manager/history"
	"scaffold/manager/project"
	"scaffold/manager/release"
	"sort"
	"strings"

	"github.com/jfcarter2358/ui"
	"github.com/jfcarter2358/ui/accordion"
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

func ReleasesTeamProjectSearchEndpoint(ctx *gin.Context) {
	p := ctx.Param("project")
	t := ctx.Param("team")
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	releases, err := release.GetReleases(bson.M{"team": t, "project": p})
	if err != nil {
		logger.Errorf("", "Cannot render releases page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	filtered := []release.Release{}

	for _, r := range releases {
		if strings.Contains(strings.ToLower(r.ID), strings.ToLower(query)) {
			filtered = append(filtered, *r)
		}
	}

	markdown := releasesTeamProjectBuildTable(filtered, t, p, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func ReleasesTeamProjectTableEndpoint(ctx *gin.Context) {
	p := ctx.Param("project")
	t := ctx.Param("team")
	releases, err := release.GetReleases(bson.M{"team": t, "project": p})
	if err != nil {
		logger.Errorf("", "Cannot render releases page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []release.Release{}

	for _, r := range releases {
		filtered = append(filtered, *r)
	}

	markdown := releasesTeamProjectBuildTable(filtered, t, p, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func ReleasesTeamProjectPageEndpoint(ctx *gin.Context) {
	markdown := releasesTeamProjectBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func releasesTeamProjectBuildPage(ctx *gin.Context) []byte {
	p := ctx.Param("project")
	t := ctx.Param("team")
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
										Title: fmt.Sprintf("%s/%s", t, p),
										HRef:  fmt.Sprintf("/ui/releases/%s/%s", t, p),
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: fmt.Sprintf(`<input id="search" class="w3-input  search-bar theme-light" type="text"
                            name="search" placeholder="Search Releases"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%%;width:98%%" hx-get="/htmx/releases/%s/%s/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#releases-table-div" />`, t, p),
					},
					div.Div{
						ID:        "releases-table-div",
						HXTrigger: "load",
						HXGet:     fmt.Sprintf("/htmx/releases/%s/%s/table", t, p),
						Style:     "margin-top:8px;margin-bottom:8px;margin-left:1%;width:98%",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := pp.Render()
	if err != nil {
		logger.Errorf("", "Cannot render releases page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

func releasesTeamProjectBuildTable(rs []release.Release, t, p string, ctx *gin.Context) []byte {

	out := ""

	es, err := project.GetEnvironments(bson.M{"team": t, "project": p})
	if err != nil {
		logger.Errorf("", "Could not get environments for %s/%s: %s", t, p, err)
		return []byte{}
	}

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Updated > rs[j].Updated
	})

	// tt := table.Table{
	// 	ID: "releases_table",
	// 	Headers: []header.Header{
	// 		{
	// 			Contents: "ID",
	// 			Classes:  "text-lg",
	// 		},
	// 		{
	// 			Contents: "Created",
	// 			Classes:  "text-lg",
	// 		},
	// 		{
	// 			Contents: "Updated",
	// 			Classes:  "text-lg",
	// 		},
	// 		{
	// 			Contents: "Status",
	// 			Classes:  "text-lg",
	// 		},
	// 		{
	// 			Contents: "",
	// 			Classes:  "text-lg",
	// 		},
	// 	},
	// 	Rows:          make([][]cell.Cell, 0),
	// 	Classes:       "theme-light",
	// 	Style:         "width:100%;",
	// 	HeaderClasses: "ui-green",
	// }

	// token, _ := ctx.Cookie("scaffold_token")
	// u, _ := user.GetUserByLoginToken(token)

	for _, r := range rs {
		rows := []ui.Component{}
		// rr := []cell.Cell{
		// 	{
		// 		Contents: r.ID,
		// 	},
		// 	{
		// 		Contents: r.Created,
		// 	},
		// 	{
		// 		Contents: r.Updated,
		// 	},
		// 	{
		// 		Contents: r.Status,
		// 	},
		// 	{
		// 		Contents: fmt.Sprintf(`<a href="/ui/releases/%s/%s/%s" class="table-link-link w3-right-align dark theme-text"
		// 		style="float:right;margin-right:16px;">
		// 		<i class="fa-solid fa-link"></i>
		// 	</a>`, t, p, r.ID),
		// 	},
		// }

		// tt.Rows = append(tt.Rows, rr)

		logger.Tracef("", "Got environments %v", es)

		for _, e := range es {
			logger.Tracef("", "Getting histories for %s/%s/%s", t, p, e.Name)
			hs, err := history.GetHistories(bson.M{"team": t, "project": p, "environment": e.Name})
			if err != nil {
				logger.Errorf("", "Cannot get history at %s/%s/%s: %s", t, p, e.Name, err)
				return []byte{}
			}
			sort.Slice(hs, func(i, j int) bool {
				return hs[i].Created > hs[j].Created
			})
			et := table.Table{
				ID: "releases_table",
				Headers: []header.Header{
					{
						Contents: "Run ID",
						Classes:  "text-lg",
					},
					{
						Contents: "Environment",
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
			logger.Tracef("", "Got histories of %v", hs)
			for _, h := range hs {
				// eRows = append(eRows, ui.Raw{
				// 	HTMLString: fmt.Sprintf(`<a href="/ui/releases/%s/%s/%s/%s" class="dark theme-base shadow-xl ui-hover-green" style="width:100%%;padding:8px;display:inline-block;">%s</a>`, t, p, e.Name, h.ReleaseID, h.RunID),
				// })
				rr := []cell.Cell{
					{
						Contents: h.RunID,
					},
					{
						Contents: h.Environment,
					},
					{
						Contents: h.Created,
					},
					{
						Contents: h.Updated,
					},
					{
						Contents: fmt.Sprintf(`<a href="/ui/releases/%s/%s/%s/%s" class="table-link-link w3-right-align dark theme-text"
						style="float:right;margin-right:16px;">
						<i class="fa-solid fa-link"></i>
					</a>`, t, p, e.Name, h.RunID),
					},
				}
				et.Rows = append(et.Rows, rr)
			}
			rows = append(rows, et)
		}
		rows = append(rows, br.BR{})
		a := accordion.Accordion{
			Classes:      "card shadow-xl theme-base theme-border-base",
			ID:           "runtime-accordion",
			ContentStyle: "padding:0px;margin-top:8px;margin-left:1%;width:98%;",
			Items: []accordion.AccordionItem{
				{
					Title:      r.Name,
					Classes:    "ui-green",
					Components: rows,
				},
			},
		}
		html, err := a.Render()
		if err != nil {
			logger.Errorf("", "Cannot render runbooks card: %s", err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return []byte{}
		}
		out += html + "</br>"
	}

	// logger.Tracef("", out)

	return []byte(out)
}
