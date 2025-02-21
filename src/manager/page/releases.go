package page

import (
	"fmt"
	"net/http"
	"scaffold/manager/release"
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

func ReleasesSearchEndpoint(ctx *gin.Context) {
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	releases, err := release.GetReleases(bson.M{})
	if err != nil {
		logger.Errorf("", "Cannot render releases page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	filtered := []release.Release{}

	for _, r := range releases {
		if strings.Contains(strings.ToLower(r.ID), strings.ToLower(query)) || strings.Contains(strings.ToLower(r.Team), strings.ToLower(query)) || strings.Contains(strings.ToLower(r.Project), strings.ToLower(query)) || strings.Contains(strings.ToLower(r.Status), strings.ToLower(query)) {
			filtered = append(filtered, *r)
		}
	}

	markdown := releasesBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func ReleasesTableEndpoint(ctx *gin.Context) {
	releases, err := release.GetReleases(bson.M{})
	if err != nil {
		logger.Errorf("", "Cannot render releases page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []release.Release{}

	for _, r := range releases {
		filtered = append(filtered, *r)
	}

	logger.Tracef("", "Got %d releases", len(releases))

	markdown := releasesBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func ReleasesPageEndpoint(ctx *gin.Context) {
	markdown := releasesBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func releasesBuildPage(ctx *gin.Context) []byte {
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
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: `<input id="search" class="w3-input search-bar theme-light" type="text"
                            name="search" placeholder="Search Releases"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%;width:98%" hx-get="/htmx/releases/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#releases-table-div" />`,
					},
					div.Div{
						ID:        "releases-table-div",
						HXTrigger: "load",
						HXGet:     "/htmx/releases/table",
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

func releasesBuildTable(rs []release.Release, ctx *gin.Context) []byte {
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Name < rs[j].Name
	})

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Project < rs[j].Project
	})

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Team < rs[j].Team
	})

	// tKeys := []string{}
	// pKeys := map[string][]string{}

	// rrs := make(map[string]map[string]release.Release)
	// for _, r := range rs {
	// 	if _, ok := rrs[r.Team]; ok {
	// 		if _, ok := rrs[r.Team][r.Project]; ok {
	// 			continue
	// 		}
	// 		rrs[r.Team][r.Project] = r
	// 		pp := pKeys[r.Team]
	// 		pp = append(pp, r.Project)
	// 		pKeys[r.Team] = pp
	// 		continue
	// 	}
	// 	rrs[r.Team] = map[string]release.Release{
	// 		r.Project: r,
	// 	}
	// 	tKeys = append(tKeys, r.Team)
	// 	pKeys[r.Team] = []string{r.Project}
	// }

	t := table.Table{
		ID: "releases_table",
		Headers: []header.Header{
			{
				Contents: "ID",
				Classes:  "text-lg",
			},
			{
				Contents: "Name",
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

	// token, _ := ctx.Cookie("scaffold_token")
	// u, _ := user.GetUserByLoginToken(token)

	// logger.Debugf("", "Got tKeys of %v", tKeys)
	// logger.Debugf("", "Got pKeys of %v", pKeys)
	// for _, t := range tKeys {
	// 	logger.Tracef("", "Got team of %s", t)
	// 	m := rrs[t]
	// 	for _, p := range pKeys[t] {
	// 		logger.Tracef("", "Got project of %s", p)
	for _, rb := range rs {
		// r := m[p]
		r := []cell.Cell{
			{
				Contents: rb.ID,
			},
			{
				Contents: rb.Name,
			},
			{
				Contents: rb.Team,
			},
			{
				Contents: rb.Project,
			},
			{
				Contents: rb.Status,
			},
			{
				Contents: fmt.Sprintf(`<a href="/ui/releases/%s" class="table-link-link w3-right-align dark theme-text"
				style="float:right;margin-right:16px;">
				<i class="fa-solid fa-link"></i>
			</a>`, rb.ID),
			},
		}
		t.Rows = append(t.Rows, r)
		// }
	}

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

func getKeysR(m map[string]release.Release) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func getKeysRRS(m map[string]map[string]release.Release) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
