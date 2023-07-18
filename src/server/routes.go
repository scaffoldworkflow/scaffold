// routes.go

package main

import (
	"net/http"
	"scaffold/server/api"
	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/middleware"
	"scaffold/server/page"

	"github.com/gin-gonic/gin"
)

func initializeRoutes() {
	router.Static("/static/css", "./static/css")
	router.Static("/static/img", "./static/img")
	router.Static("/static/js", "./static/js")

	router.GET("/", page.RedirectIndexPage)

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{})
	})

	healthRoutes := router.Group("/health", middleware.CORSMiddleware())
	{
		healthRoutes.GET("/healthy", api.Healthy)
		healthRoutes.GET("/ready", api.Ready)
		if config.Config.Node.Type == constants.NODE_TYPE_WORKER {
			healthRoutes.GET("/available", api.Available)
		}
	}

	if config.Config.Node.Type == constants.NODE_TYPE_MANAGER {
		authRoutes := router.Group("/auth", middleware.CORSMiddleware())
		{
			authRoutes.POST("/login", middleware.EnsureNotLoggedIn(), auth.PerformLogin)
			authRoutes.GET("/logout", middleware.EnsureLoggedIn(), auth.PerformLogout)
			authRoutes.POST("/reset/request", middleware.EnsureNotLoggedIn(), auth.RequestPasswordReset)
			authRoutes.POST("/reset/do", middleware.EnsureNotLoggedIn(), auth.DoPasswordReset)
		}

		apiRoutes := router.Group("/api", middleware.CORSMiddleware())
		{
			v1Routes := apiRoutes.Group("/v1")
			{
				cascadeRoutes := v1Routes.Group("/cascade")
				{
					cascadeRoutes.GET("", api.GetAllCascades)
					cascadeRoutes.GET("/:name", api.GetCascadeByName)
					cascadeRoutes.DELETE("/:name", api.DeleteCascadeByName)
					cascadeRoutes.POST("", api.CreateCascade)
					cascadeRoutes.PUT("/:name", api.UpdateCascadeByName)
				}
				datastoreRoutes := v1Routes.Group("/datastore")
				{
					datastoreRoutes.GET("", api.GetAllDataStores)
					datastoreRoutes.GET("/:name", api.GetDataStoreByName)
					datastoreRoutes.DELETE("/:name", api.DeleteDataStoreByName)
					datastoreRoutes.POST("", api.CreateDataStore)
					datastoreRoutes.PUT("/:name", api.UpdateDataStoreByName)
				}
				stateRoutes := v1Routes.Group("/state")
				{
					stateRoutes.GET("", api.GetAllStates)
					stateRoutes.GET("/:name", api.GetStateByName)
					stateRoutes.DELETE("/:name", api.DeleteStateByName)
					stateRoutes.POST("", api.CreateState)
					stateRoutes.PUT("/:name", api.UpdateStateByName)
				}
				userRoutes := v1Routes.Group("/user")
				{
					userRoutes.GET("", api.GetAllUsers)
					userRoutes.GET("/:username", api.GetUserByUsername)
					userRoutes.DELETE("/:username", api.DeleteUserByUsername)
					userRoutes.POST("", api.CreateUser)
					userRoutes.PUT("/:username", api.UpdateUserByUsername)
				}
			}
		}

		uiRoutes := router.Group("/ui", middleware.CORSMiddleware())
		{
			uiRoutes.GET("/login", middleware.EnsureNotLoggedIn(), page.ShowLoginPage)
			uiRoutes.GET("/forgot_password", middleware.EnsureNotLoggedIn(), page.ShowForgotPasswordPage)
			uiRoutes.GET("/email_success", middleware.EnsureNotLoggedIn(), page.ShowEmailSuccessPage)
			uiRoutes.GET("/email_failure", middleware.EnsureNotLoggedIn(), page.ShowEmailFailurePage)
			uiRoutes.GET("/reset_password/:reset_password", middleware.EnsureNotLoggedIn(), page.ShowResetPasswordPage)

			uiRoutes.GET("/cascades", middleware.EnsureLoggedIn(), page.ShowCascadesPage)
			uiRoutes.GET("/cascades/:name", middleware.EnsureLoggedIn(), page.ShowCascadePage)

			uiRoutes.GET("/users", middleware.EnsureLoggedIn(), page.ShowUsersPage)
			uiRoutes.GET("/user/:username", middleware.EnsureLoggedIn(), page.ShowUserPage)
		}
	}

	// TODO: Add worker routes
}
