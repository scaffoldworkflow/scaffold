// routes.go

package main

import (
	"scaffold/server/api"
	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/docs"
	"scaffold/server/middleware"
	"scaffold/server/page"
	"scaffold/server/page/common"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func initializeRoutes() {
	router.Static("/static/css", "./static/css")
	router.Static("/static/img", "./static/img")
	router.Static("/static/js", "./static/js")

	// Swagger docs
	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.GET("/", page.RedirectIndexPage)

	router.NoRoute(func(c *gin.Context) {
		common.Code404Endpoint(c)
	})

	healthRoutes := router.Group("/health", middleware.CORSMiddleware())
	{
		healthRoutes.GET("/healthy", api.Healthy)
		healthRoutes.GET("/ready", api.Ready)
		if config.Config.Node.Type == constants.NODE_TYPE_WORKER {
			healthRoutes.GET("/available", api.Available)
		} else {
			healthRoutes.POST("/ping/:name", middleware.EnsureLoggedIn(), api.Ping)
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
				workflowRoutes := v1Routes.Group("/workflow")
				{
					workflowRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllWorkflows)
					workflowRoutes.GET("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("name"), api.GetWorkflowByName)
					workflowRoutes.DELETE("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("name"), api.DeleteWorkflowByName)
					workflowRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateWorkflow)
					workflowRoutes.PUT("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("name"), api.UpdateWorkflowByName)
				}
				datastoreRoutes := v1Routes.Group("/datastore")
				{
					datastoreRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllDataStores)
					datastoreRoutes.GET("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("name"), api.GetDataStoreByName)
					datastoreRoutes.DELETE("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("name"), api.DeleteDataStoreByWorkflow)
					datastoreRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateDataStore)
					datastoreRoutes.PUT("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("name"), api.UpdateDataStoreByWorkflow)
				}
				fileRoutes := v1Routes.Group("/file")
				{
					fileRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllFiles)
					fileRoutes.GET("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("name"), api.GetFilesByWorkflow)
					fileRoutes.GET("/:name/:file", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("name"), api.GetFileByNames)
					// fileRoutes.GET("/:name/:file/view", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("name"), api.ViewFile)
					fileRoutes.GET("/:name/:file/download", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("name"), api.DownloadFile)
					fileRoutes.POST("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("name"), api.UploadFile)
				}
				stateRoutes := v1Routes.Group("/state")
				{
					stateRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllStates)
					stateRoutes.GET("/:workflow", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("workflow"), api.GetStatesByWorkflow)
					stateRoutes.GET("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("workflow"), api.GetStateByNames)
					stateRoutes.DELETE("/:workflow", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.DeleteStatesByWorkflow)
					stateRoutes.DELETE("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.DeleteStateByNames)
					stateRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateState)
					stateRoutes.PUT("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.UpdateStateByNames)
				}
				inputRoutes := v1Routes.Group("/input")
				{
					inputRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllInputs)
					inputRoutes.GET("/:workflow", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("workflow"), api.GetInputsByWorkflow)
					inputRoutes.GET("/:workflow/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("workflow"), api.GetInputByNames)
					inputRoutes.DELETE("/:workflow", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.DeleteInputsByWorkflow)
					inputRoutes.DELETE("/:workflow/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.DeleteInputByNames)
					inputRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateInput)
					inputRoutes.POST("/:workflow/update", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.UpdateInputDependenciesByName)
					inputRoutes.PUT("/:workflow/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.UpdateInputByNames)
				}
				taskRoutes := v1Routes.Group("/task")
				{
					taskRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllTasks)
					taskRoutes.GET("/:workflow", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("workflow"), api.GetTasksByWorkflow)
					taskRoutes.GET("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureWorkflowGroup("workflow"), api.GetTaskByNames)
					taskRoutes.DELETE("/:workflow", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.DeleteTasksByWorkflow)
					taskRoutes.DELETE("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.DeleteTaskByNames)
					taskRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateTask)
					taskRoutes.PUT("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.UpdateTaskByNames)
					taskRoutes.PUT("/:workflow/:task/enabled", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.ToggleTaskEnabled)
				}
				userRoutes := v1Routes.Group("/user")
				{
					userRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllUsers)
					userRoutes.GET("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetUserByUsername)
					userRoutes.DELETE("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteUserByUsername)
					userRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin"}), api.CreateUser)
					userRoutes.PUT("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateUserByUsername)
				}
				runRoutes := v1Routes.Group("/run")
				{
					runRoutes.POST("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.CreateRun)
					runRoutes.DELETE("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureWorkflowGroup("workflow"), api.ManagerKillRun)
				}
				webhookRoutes := v1Routes.Group("/webhook")
				{
					webhookRoutes.POST("/:workflow/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureWorkflowGroup("workflow"), api.TriggerWebhookByID)
				}
			}
		}

		uiRoutes := router.Group("/ui", middleware.CORSMiddleware())
		{
			uiRoutes.GET("/login", middleware.EnsureNotLoggedIn(), page.LoginPageEndpoint)
			uiRoutes.GET("/forgot_password", middleware.EnsureNotLoggedIn(), page.ShowForgotPasswordPage)
			uiRoutes.GET("/email_success", middleware.EnsureNotLoggedIn(), page.ShowEmailSuccessPage)
			uiRoutes.GET("/email_failure", middleware.EnsureNotLoggedIn(), page.ShowEmailFailurePage)
			uiRoutes.GET("/reset_password/:reset_password", middleware.EnsureNotLoggedIn(), page.ShowResetPasswordPage)

			uiRoutes.GET("/dashboard", middleware.EnsureLoggedIn(), page.DashboardPageEndpoint)

			uiRoutes.GET("/workflows", middleware.EnsureLoggedIn(), page.WorkflowsPageEndpoint)
			uiRoutes.GET("/workflows/:name", middleware.EnsureLoggedIn(), page.WorkflowPageEndpoint)

			uiRoutes.GET("/runs", middleware.EnsureLoggedIn(), page.HistoriesPageEndpoint)
			uiRoutes.GET("/runs/:run_id", middleware.EnsureLoggedIn(), page.HistoryPageEndpoint)

			uiRoutes.GET("/users", middleware.EnsureLoggedIn(), page.UsersPageEndpoint)
			uiRoutes.GET("/users/:username", middleware.EnsureLoggedIn(), page.UserPageEndpoint)

			uiRoutes.GET("/401", common.Code401Endpoint)
			uiRoutes.GET("/403", common.Code403Endpoint)
			uiRoutes.GET("/404", common.Code404Endpoint)
			uiRoutes.GET("/500", common.Code500Endpoint)
		}

		htmxRoutes := router.Group("/htmx", middleware.CORSMiddleware(), middleware.EnsureLoggedInAPI())
		{
			// commonRoutes := htmxRoutes.Group("/common")
			// {
			// 	if err := common.Init(); err != nil {
			// 		panic(err)
			// 	}
			// 	commonRoutes.GET("/status", common.StatusEndpoint)
			// 	commonRoutes.GET("/sidebar", common.SidebarEndpoint)
			// 	commonRoutes.GET("/error", common.ErrorEndpoint)
			// 	commonRoutes.GET("/success", common.ErrorEndpoint)
			// 	commonRoutes.GET("/header", common.HeaderEndpoint)
			// }
			workflowsRoutes := htmxRoutes.Group("/workflows")
			{
				workflowsRoutes.GET("/table", page.WorkflowsTableEndpoint)
				workflowsRoutes.GET("/search", page.WorkflowsSearchEndpoint)
			}
			dashboardRoutes := htmxRoutes.Group("/dashboard")
			{
				dashboardRoutes.GET("/table", page.DashboardTableEndpoint)
				dashboardRoutes.GET("/search", page.DashboardSearchEndpoint)
			}
			runsRoutes := htmxRoutes.Group("/runs")
			{
				runsRoutes.GET("/table", page.HistoriesTableEndpoint)
				runsRoutes.GET("/search", page.HistoriesSearchEndpoint)
				runsRoutes.GET("/timeline/:run_id", page.HistoryTimelineEndpoint)
				runsRoutes.GET("/timeline/:run_id/status/:state_name", page.HistoryStateEndpoint)
			}
			usersRoutes := htmxRoutes.Group("/users")
			{
				usersRoutes.GET("/table", page.UsersTableEndpoint)
				usersRoutes.GET("/search", page.UsersSearchEndpoint)
			}
		}

	}
	if config.Config.Node.Type == constants.NODE_TYPE_WORKER {
		apiRoutes := router.Group("/api", middleware.CORSMiddleware())
		{
			v1Routes := apiRoutes.Group("/v1")
			{
				taskRoutes := v1Routes.Group("/run")
				{
					taskRoutes.DELETE("/:workflow/:task", middleware.EnsureLoggedIn(), api.KillRun)
				}
			}
		}
	}
}
