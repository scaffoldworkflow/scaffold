package page

// import (
// 	"fmt"
// 	"net/http"
// 	"scaffold/manager/history"
// 	"sort"

// 	"github.com/jfcarter2358/ui"
// 	"github.com/jfcarter2358/ui/breadcrumb"
// 	"github.com/jfcarter2358/ui/card"
// 	"github.com/jfcarter2358/ui/collapse"
// 	"github.com/jfcarter2358/ui/elements/br"
// 	"github.com/jfcarter2358/ui/elements/div"
// 	"github.com/jfcarter2358/ui/elements/h1"
// 	"github.com/jfcarter2358/ui/elements/link"
// 	"github.com/jfcarter2358/ui/elements/pre"
// 	"github.com/jfcarter2358/ui/modal"
// 	"github.com/jfcarter2358/ui/page"
// 	"github.com/jfcarter2358/ui/topbar"
// 	"go.mongodb.org/mongo-driver/bson"

// 	_ "embed"

// 	"github.com/gin-gonic/gin"
// 	logger "github.com/jfcarter2358/go-logger"
// )

// func ServicePageEndpoint(ctx *gin.Context) {
// 	markdown := releaseBuildPage(ctx)
// 	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
// }

// func serviceBuildPage(ctx *gin.Context) []byte {
// 	pName := ctx.Param("project")
// 	tName := ctx.Param("team")
// 	eName := ctx.Param("environment")
// 	sName := ctx.Param("service")
// 	rID := ctx.Param("run_id")

// 	w := releaseBuildWorkflow(tName, pName, eName, rID, ctx)
// 	p := page.Page{
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
// 										Title: fmt.Sprintf("%s/%s", tName, pName),
// 										HRef:  fmt.Sprintf("/ui/releases/%s/%s", tName, pName),
// 									},
// 									link.Link{
// 										Title: fmt.Sprintf("%s/%s", eName, rID),
// 										HRef:  fmt.Sprintf("/ui/releases/%s/%s/%s/%s", tName, pName, eName, rID),
// 									},
// 								},
// 								Style: "margin-left:16px;",
// 							},
// 						},
// 					},
// 					div.Div{
// 						ID:    "workflow-div",
// 						Style: "padding-left:64px;",
// 						Components: []ui.Component{
// 							ui.Raw{
// 								HTMLString: string(w),
// 							},
// 						},
// 					},
// 				},
// 				Style: "margin:64px;",
// 			},
// 			br.BR{},
// 		},
// 	}
// 	html, err := p.Render()
// 	if err != nil {
// 		logger.Errorf("", "Cannot render run page: %s", err.Error())
// 		ctx.AbortWithStatus(http.StatusInternalServerError)
// 		return []byte{}
// 	}
// 	return []byte(html)
// }

// func releaseBuildWorkflow(t, p, e, rID string, ctx *gin.Context) []byte {
// 	hs, err := history.GetHistories(bson.M{"team": t, "project": p, "environment": e, "run_id": rID})
// 	if err != nil {
// 		logger.Errorf("", "Unable to get environment for %s/%s/%s: %s", t, p, e, err)
// 		return []byte{}
// 	}
// 	sort.Slice(hs, func(i, j int) bool {
// 		return hs[i].Updated > hs[j].Updated
// 	})
// 	if len(hs) > 0 {
// 		d := div.Div{
// 			Classes: "theme-light",
// 			Components: []ui.Component{
// 				card.Card{
// 					ID:      "workflow-card",
// 					Classes: "theme-light",
// 					Style:   "width:100%;padding:0px;height:100%;",
// 				},
// 				modal.Modal{
// 					ID: "state_modal",
// 					Components: []ui.Component{
// 						h1.H1{
// 							Contents: `<h1 id="current-state-header" class="text-3xl" style="float:left;padding-top:8px;"></h1>`,
// 							Classes:  "ui-green",
// 							Style:    "width:100%;",
// 						},
// 						br.BR{},
// 						br.BR{},
// 						br.BR{},
// 						collapse.Collapse{
// 							TitleID:      "state-status-collapse",
// 							Classes:      "theme-light",
// 							TitleClasses: "ui-green",
// 							Title:        "Status",
// 							Components: []ui.Component{
// 								br.BR{},
// 								div.Div{
// 									Components: []ui.Component{
// 										ui.Raw{
// 											HTMLString: `<span id="state-status"></span>`,
// 										},
// 										br.BR{},
// 										ui.Raw{
// 											HTMLString: `<span id="state-started"></span>`,
// 										},
// 										br.BR{},
// 										ui.Raw{
// 											HTMLString: `<span id="state-finished"></span>`,
// 										},
// 										br.BR{},
// 									},
// 								},
// 							},
// 						},
// 						br.BR{},
// 						collapse.Collapse{
// 							TitleID:      "state-context-collapse",
// 							Classes:      "theme-light",
// 							TitleClasses: "ui-green",
// 							Title:        "Context",
// 							Components: []ui.Component{
// 								br.BR{},
// 								div.Div{
// 									ID: "state-context",
// 								},
// 							},
// 						},
// 						br.BR{},
// 						collapse.Collapse{
// 							TitleID:      "state-display-collapse",
// 							Classes:      "theme-light",
// 							TitleClasses: "ui-green",
// 							Title:        "Display",
// 							Components: []ui.Component{
// 								div.Div{
// 									ID: "state-current-display-data",
// 								},
// 							},
// 						},
// 						br.BR{},
// 						collapse.Collapse{
// 							TitleID:      "state-output-collapse",
// 							Classes:      "theme-light",
// 							TitleClasses: "ui-green",
// 							Title:        "Output",
// 							Components: []ui.Component{
// 								pre.Pre{
// 									ID:    "state-output",
// 									Style: "font-family:monospace;overflow-x:scroll",
// 								},
// 							},
// 						},
// 					},
// 					BoxClasses: "max-w-none w-4/5",
// 				},
// 				ui.Raw{
// 					HTMLString: `
// 							<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.5.1/jquery.min.js"></script>
// 							<script src="https://malsup.github.io/jquery.form.js"></script>
// 							<script src="/static/js/jquery-ui.min.js"></script>
// 							<script src="/static/js/dagre.min.js"></script>
// 							<script src="/static/js/workflow.js"></script>
// 							`,
// 				},
// 			},
// 		}

// 		html, err := d.Render()
// 		if err != nil {
// 			logger.Errorf("", "Cannot render run workflow: %s", err.Error())
// 			ctx.AbortWithStatus(http.StatusInternalServerError)
// 			return []byte{}
// 		}
// 		return []byte(html)
// 	}

// 	return []byte{}
// }
