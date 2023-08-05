// routes.go

package main

import (
	"net/http"
	"scaffold/server/api"
	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/manager"
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
		} else {
			healthRoutes.GET("/status", middleware.EnsureLoggedIn(), manager.GetStatus)
		}
	}

	if config.Config.Node.Type == constants.NODE_TYPE_MANAGER {
		authRoutes := router.Group("/auth", middleware.CORSMiddleware())
		{
			authRoutes.POST("/login", middleware.EnsureNotLoggedIn(), auth.PerformLogin)
			authRoutes.GET("/logout", middleware.EnsureLoggedIn(), auth.PerformLogout)
			authRoutes.POST("/reset/request", middleware.EnsureNotLoggedIn(), auth.RequestPasswordReset)
			authRoutes.POST("/reset/do", middleware.EnsureNotLoggedIn(), auth.DoPasswordReset)
			authRoutes.POST("/join", auth.JoinNode)
			authRoutes.POST("/token/:username/:name", middleware.EnsureLoggedIn(), middleware.EnsureSelf(), api.GenerateAPIToken)
			authRoutes.DELETE("/token/:username/:name", middleware.EnsureLoggedIn(), middleware.EnsureSelf(), api.RevokeAPIToken)
		}

		apiRoutes := router.Group("/api", middleware.CORSMiddleware())
		{
			v1Routes := apiRoutes.Group("/v1")
			{
				cascadeRoutes := v1Routes.Group("/cascade")
				{
					cascadeRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllCascades)
					cascadeRoutes.GET("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetCascadeByName)
					cascadeRoutes.DELETE("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteCascadeByName)
					cascadeRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateCascade)
					cascadeRoutes.PUT("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateCascadeByName)
				}
				datastoreRoutes := v1Routes.Group("/datastore")
				{
					datastoreRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllDataStores)
					datastoreRoutes.GET("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetDataStoreByName)
					datastoreRoutes.DELETE("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteDataStoreByName)
					datastoreRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateDataStore)
					datastoreRoutes.PUT("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateDataStoreByName)
					datastoreRoutes.GET("/file/:name/:file", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.DownloadFile)
					datastoreRoutes.POST("/file/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.UploadFile)
				}
				stateRoutes := v1Routes.Group("/state")
				{
					stateRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllStates)
					stateRoutes.GET("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetStatesByCascade)
					stateRoutes.GET("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetStateByNames)
					stateRoutes.DELETE("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteStatesByCascade)
					stateRoutes.DELETE("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteStateByNames)
					stateRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateState)
					stateRoutes.PUT("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateStateByNames)
				}
				inputRoutes := v1Routes.Group("/input")
				{
					inputRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllInputs)
					inputRoutes.GET("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetInputsByCascade)
					inputRoutes.GET("/:cascade/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetInputByNames)
					inputRoutes.DELETE("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteInputsByCascade)
					inputRoutes.DELETE("/:cascade/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteInputByNames)
					inputRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateInput)
					inputRoutes.POST("/:cascade/update", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateInputDependenciesByName)
					inputRoutes.PUT("/:cascade/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateInputByNames)
				}
				taskRoutes := v1Routes.Group("/task")
				{
					taskRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllTasks)
					taskRoutes.GET("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetTasksByCascade)
					taskRoutes.GET("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetTaskByNames)
					taskRoutes.DELETE("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteTasksByCascade)
					taskRoutes.DELETE("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteTaskByNames)
					taskRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateTask)
					taskRoutes.PUT("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateTaskByNames)
				}
				userRoutes := v1Routes.Group("/user")
				{
					userRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllUsers)
					userRoutes.GET("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetUserByUsername)
					userRoutes.DELETE("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteUserByUsername)
					userRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateUser)
					userRoutes.PUT("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateUserByUsername)
				}
				runRoutes := v1Routes.Group("/run")
				{
					runRoutes.POST(":cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateRun)
					runRoutes.POST(":cascade/:task/check", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateCheckRun)
					runRoutes.GET("/containers", middleware.EnsureLoggedIn(), api.GetAllContainers)
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

			uiRoutes.GET("/files", middleware.EnsureLoggedIn(), page.ShowFilesPage)

			uiRoutes.GET("/users", middleware.EnsureLoggedIn(), page.ShowUsersPage)
			uiRoutes.GET("/user/:username", middleware.EnsureLoggedIn(), page.ShowUserPage)
		}
	}
	if config.Config.Node.Type == constants.NODE_TYPE_WORKER {
		apiRoutes := router.Group("/api", middleware.CORSMiddleware())
		{
			v1Routes := apiRoutes.Group("/v1")
			{
				v1Routes.POST("/trigger", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.TriggerRun)
				v1Routes.GET("/state/:cascade/:task/:number", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetRunState)
				v1Routes.GET("/available", middleware.EnsureLoggedIn(), api.GetAvailableContainers)
			}
		}
	}
}
