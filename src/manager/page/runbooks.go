package page

import (
	"fmt"
	"net/http"
	"scaffold/manager/runbook"
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

func RunbooksSearchEndpoint(ctx *gin.Context) {
	searchTerm, ok := ctx.GetQuery("search")
	if !ok {
		ctx.Status(http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(searchTerm)

	runbooks, err := runbook.GetRunbooks(bson.M{})
	if err != nil {
		logger.Errorf("", "Cannot render runbooks page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	filtered := []runbook.Runbook{}

	for _, r := range runbooks {
		if strings.Contains(strings.ToLower(r.Team), strings.ToLower(query)) || strings.Contains(strings.ToLower(r.Name), strings.ToLower(query)) || strings.Contains(strings.ToLower(r.Project), strings.ToLower(query)) {
			filtered = append(filtered, *r)
		}
	}

	markdown := runbooksBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func RunbooksTableEndpoint(ctx *gin.Context) {
	runbooks, err := runbook.GetRunbooks(bson.M{})
	if err != nil {
		logger.Errorf("", "Cannot render runbooks page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	filtered := []runbook.Runbook{}

	logger.Tracef("", "Got runbooks: %v", runbooks)

	for _, r := range runbooks {
		filtered = append(filtered, *r)
	}

	markdown := runbooksBuildTable(filtered, ctx)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func RunbooksPageEndpoint(ctx *gin.Context) {
	markdown := runbooksBuildPage(ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func runbooksBuildPage(ctx *gin.Context) []byte {
	p := page.Page{
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
										Title: "Runbooks",
										HRef:  "/ui/runbooks",
									},
								},
								Style: "margin-left:16px;",
							},
						},
					},
					ui.Raw{
						HTMLString: `<input id="search" class="w3-input search-bar theme-light" type="text"
                            name="search" placeholder="Search Runbooks"
                            style="margin-top:8px;margin-bottom:8px;margin-left:1%%;width:98%%" hx-get="/htmx/runbooks/search"
                            hx-trigger="keyup changed delay:250ms" hx-target="#runbooks-table-div" />`,
					},
					div.Div{
						ID:        "runbooks-table-div",
						HXTrigger: "load",
						HXGet:     "/htmx/runbooks/table",
					},
				},
				Style: "margin:64px;",
			},
			br.BR{},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render runbooks page: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}

// func runbooksBuildTable(rs []runbook.Runbook, ctx *gin.Context) []byte {
// 	token, _ := ctx.Cookie("scaffold_token")
// 	u, _ := user.GetUserByLoginToken(token)

// 	out := ""

// 	categorized := map[string][]runbook.Runbook{}

// 	for _, r := range rs {
// 		if val, ok := categorized[r.Category]; ok {
// 			categorized[r.Category] = append(val, r)
// 		} else {
// 			categorized[r.Category] = []runbook.Runbook{r}
// 		}
// 	}

// 	logger.Debugf("", "Got categorized of %v", categorized)

// 	for category, arr := range categorized {
// 		sort.Slice(arr, func(i, j int) bool {
// 			return arr[i].Name < arr[j].Name
// 		})

// 		logger.Debugf("", "Got arr of %v in category %s", arr, category)

// 		rows := []ui.Component{}

// 		for _, r := range arr {
// 			logger.Tracef("", "Checking %v against %v", u.Groups, r.Groups)
// 			isInGroup := false
// 			for _, ug := range u.Groups {
// 				if ug == "admin" {
// 					isInGroup = true
// 					break
// 				}
// 				for _, rg := range r.Groups {
// 					if ug == rg {
// 						isInGroup = true
// 						break
// 					}
// 				}
// 				if isInGroup {
// 					break
// 				}
// 			}
// 			if isInGroup {
// 				logger.Tracef("", "Auth confirmed")
// 				rows = append(rows, ui.Raw{
// 					HTMLString: `<a href="/ui/runbooks/` + r.ID + `" class="dark theme-base shadow-xl" style="width:100%;padding:8px;display:inline-block;">` + r.Name + `</a>`,
// 				})
// 			}
// 		}
// 		a := accordion.Accordion{
// 			Classes:      "card shadow-xl theme-base theme-border-base",
// 			ID:           "runtime-accordion",
// 			ContentStyle: "padding:0px;",
// 			Items: []accordion.AccordionItem{
// 				{
// 					Title:      category,
// 					Classes:    "ui-green",
// 					Components: rows,
// 				},
// 			},
// 		}
// 		html, err := a.Render()
// 		if err != nil {
// 			logger.Errorf("", "Cannot render runbooks card: %s", err.Error())
// 			ctx.AbortWithStatus(http.StatusInternalServerError)
// 			return []byte{}
// 		}
// 		out += html + "</br>"
// 	}

// 	return []byte(out)
// }

func runbooksBuildTable(rs []runbook.Runbook, ctx *gin.Context) []byte {
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Name < rs[j].Name
	})

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Project < rs[j].Project
	})
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Team < rs[j].Team
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
				Contents: "Project",
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

	for _, rb := range rs {
		r := []cell.Cell{
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
				Contents: rb.Created,
			},
			{
				Contents: rb.Updated,
			},
			{
				Contents: fmt.Sprintf(`<a href="/ui/runbooks/%s" class="table-link-link w3-right-align dark theme-text"
							style="float:right;margin-right:16px;">
							<i class="fa-solid fa-link"></i>
						</a>`, rb.ID),
			},
		}
		t.Rows = append(t.Rows, r)
	}

	html, err := t.Render()
	if err != nil {
		logger.Errorf("", "Cannot render runbooks table: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
