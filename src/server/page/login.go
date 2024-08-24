package page

import (
	"net/http"
	"scaffold/server/constants"
	"scaffold/server/ui"
	"scaffold/server/ui/elements/br"
	"scaffold/server/ui/page"

	"github.com/gin-gonic/gin"
	logger "github.com/jfcarter2358/go-logger"
)

func LoginPageEndpoint(ctx *gin.Context) {
	markdown := loginBuildPage("", ctx)
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", markdown)
}

func loginBuildPage(errorMessage string, ctx *gin.Context) []byte {
	p := page.Page{
		ID: "page",
		Components: []ui.Component{
			br.BR{},
			br.BR{},
			br.BR{},
			br.BR{},
			ui.Raw{
				HTMLString: `
<div class="modal-content dark theme-light animate w3-border w3-border-black w3-card w3-round" style="width:40%;margin-left:30%;margin-right:30%;">
<div class="w3-container scaffold-green w3-round">
	<div class="w3-center">
		<h1 class="header-text"><b>Scaffold</b></h1>
	</div>
</div>
<br>
<form class="w3-container" action="/auth/login" id="login_form" method="POST">
	<div class="form-group">
		<label for="username" class="label-text scaffold-text-green">Username</label>
		<input type="text" class="form-control w3-round w3-input dark theme-light" id="username" name="username" placeholder="Username">
	</div>
	<div class="form-group">
		<label for="password" class="label-text scaffold-text-green">Password</label>
		<input type="password" class="form-control w3-round w3-input dark theme-light" id="password" name="password"
			placeholder="Password">
	</div>
	<br>
	<p class="scaffold-text-red">` +
					errorMessage +
					`</p>
	<br>
	<label class="container scaffold-text-green">Remember Me
		<input type="checkbox" id="remember_me" name="remember_me" >
		<span class="checkmark"></span>
	</label>
	<br>
	<div>
		<button type="submit" class="w3-button scaffold-green w3-round diagonal-shadow-grey"><b>Login</b></button>
		<a href="/ui/forgot_password" style="padding-left:8px;">
			<div class="w3-button dark theme-light scaffold-text-green w3-round scaffold-border-grey w3-border"><b>Forgot Password</b></div>
		</a>
	</div>
</form>
<br>
<div class="w3-container scaffold-green w3-round">
	<p style="float: right;" class="footer-text w3-text-white"><b>v` + constants.VERSION + `</b></p>
</div>
</div>`,
			},
		},
	}
	html, err := p.Render()
	if err != nil {
		logger.Errorf("", "Cannot render login page : %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return []byte{}
	}
	return []byte(html)
}
