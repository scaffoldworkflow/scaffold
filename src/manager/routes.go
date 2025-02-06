// routes.go

package main

import (
	"scaffold/manager/api"
	"scaffold/manager/auth"
	"scaffold/manager/config"
	"scaffold/manager/constants"
	"scaffold/manager/middleware"
	"scaffold/manager/page"
	// "scaffold/manager/page/common"
)

func initializeRoutes() {
	router.Static("/static/css", "./static/css")
	router.Static("/static/img", "./static/img")
	router.Static("/static/js", "./static/js")

	// Swagger docs
	// docs.SwaggerInfo.BasePath = "/api/v1"
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// router.GET("/", page.RedirectIndexPage)

	// if err := common.Init(); err != nil {
	// 	panic(err)
	// }

	// router.NoRoute(func(c *gin.Context) {
	// 	common.Code404Endpoint(c)
	// })

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
	authRoutes := router.Group("/auth", middleware.CORSMiddleware())
	{
		authRoutes.POST("/login", middleware.EnsureNotLoggedIn(), auth.PerformLogin)
		authRoutes.GET("/logout", middleware.EnsureLoggedIn(), auth.PerformLogout)
		authRoutes.POST("/reset/request", middleware.EnsureNotLoggedIn(), auth.RequestPasswordReset)
		authRoutes.POST("/reset/do", middleware.EnsureNotLoggedIn(), auth.DoPasswordReset)
		authRoutes.POST("/join", auth.JoinNode)
		authRoutes.POST("/token/:username/:name", middleware.EnsureLoggedIn(), api.GenerateAPIToken)
		authRoutes.DELETE("/token/:username/:name", middleware.EnsureLoggedIn(), api.RevokeAPIToken)
	}

	apiRoutes := router.Group("/api", middleware.CORSMiddleware())
	{
		v1Routes := apiRoutes.Group("/v1")
		{
			projectRoutes := v1Routes.Group("/project")
			{
				projectRoutes.GET("", middleware.EnsureLoggedIn(), api.GetProjects)
				projectRoutes.DELETE("", middleware.EnsureLoggedIn(), api.DeleteProjects)
				projectRoutes.POST("", middleware.EnsureLoggedIn(), api.CreateProject)
				projectRoutes.PUT("", middleware.EnsureLoggedIn(), api.UpdateProject)
			}
			userRoutes := v1Routes.Group("/user")
			{
				userRoutes.GET("", middleware.EnsureLoggedIn(), api.GetAllUsers)
				userRoutes.GET("/:username", middleware.EnsureLoggedIn(), api.GetUserByUsername)
				userRoutes.DELETE("/:username", middleware.EnsureLoggedIn(), api.DeleteUserByUsername)
				userRoutes.POST("", middleware.EnsureLoggedIn(), api.CreateUser)
				userRoutes.PUT("/:username", middleware.EnsureLoggedIn(), api.UpdateUserByUsername)
			}
			runRoutes := v1Routes.Group("/run")
			{
				runRoutes.POST("/:project/:environment/:service/:step", middleware.EnsureLoggedIn(), api.StartRun)
				runRoutes.DELETE("/:run_id/:step", middleware.EnsureLoggedIn(), api.KillRun)
				// runRoutes.GET("/:run_id", middleware.EnsureLoggedIn(), api.GetRunStatus)
				// runRoutes.GET("/:run_id/next", middleware.EnsureLoggedIn(), api.GetRunNext)
			}
			runbookRoutes := v1Routes.Group("/runbook")
			{
				runbookRoutes.GET("", middleware.EnsureLoggedIn(), api.GetAllRunbooks)
				runbookRoutes.GET("/:id", middleware.EnsureLoggedIn(), api.GetRunbookByID)
				runbookRoutes.DELETE("/:id", middleware.EnsureLoggedIn(), api.DeleteRunbookByID)
				runbookRoutes.POST("", middleware.EnsureLoggedIn(), api.CreateRunbook)
				runbookRoutes.PUT("/:id", middleware.EnsureLoggedIn(), api.UpdateRunbookByID)
			}
			kernelRoutes := v1Routes.Group("/kernel")
			{
				kernelRoutes.GET("/:kernel_id/:run_id", middleware.EnsureLoggedIn(), api.GetKernelOutput)
				kernelRoutes.POST("/:kernel_id", middleware.EnsureLoggedIn(), api.ExecuteKernel)
			}
			// alertRoutes := v1Routes.Group("/alert")
			// {
			// 	alertRoutes.GET("", middleware.EnsureLoggedIn(), api.GetAllAlerts)
			// 	alertRoutes.GET("/:id", middleware.EnsureLoggedIn(), api.GetAlertByID)
			// 	alertRoutes.GET("/:id/workflow", middleware.EnsureLoggedIn(), api.GetAlertsByWorkflow)
			// 	alertRoutes.DELETE("/:id", middleware.EnsureLoggedIn(), api.DeleteAlertByID)
			// 	alertRoutes.DELETE("/:id/workflow", middleware.EnsureLoggedIn(), api.DeleteAlertsByWorkflow)
			// 	alertRoutes.POST("", middleware.EnsureLoggedIn(), api.CreateAlert)
			// 	alertRoutes.PUT("/:id", middleware.EnsureLoggedIn(), api.UpdateAlertByID)
			// 	alertRoutes.POST(":id/kernel", middleware.EnsureLoggedIn(), api.StartAlert)
			// 	alertRoutes.DELETE("/:id/kernel", middleware.EnsureLoggedIn(), api.StopAlert)
			// }
			historyRoutes := v1Routes.Group("/history")
			{
				historyRoutes.GET("/:runID", middleware.EnsureLoggedIn(), api.GetHistory)
				historyRoutes.GET("", middleware.EnsureLoggedIn(), api.GetAllHistories)
			}
			releaseRoutes := v1Routes.Group("/release")
			{
				releaseRoutes.GET("", middleware.EnsureLoggedIn(), api.GetReleases)
				releaseRoutes.DELETE("", middleware.EnsureLoggedIn(), api.DeleteReleases)
				releaseRoutes.POST("", middleware.EnsureLoggedIn(), api.CreateRelease)
				releaseRoutes.PUT("", middleware.EnsureLoggedIn(), api.UpdateRelease)
				releaseRoutes.POST("/release_id", middleware.EnsureLoggedIn(), api.PromoteRelease)
			}
		}
	}

	uiRoutes := router.Group("/ui", middleware.CORSMiddleware())
	{
		// 	uiRoutes.GET("/login", middleware.EnsureNotLoggedIn(), page.LoginPageEndpoint)
		// 	uiRoutes.GET("/forgot_password", middleware.EnsureNotLoggedIn(), page.ShowForgotPasswordPage)
		// 	uiRoutes.GET("/email_success", middleware.EnsureNotLoggedIn(), page.ShowEmailSuccessPage)
		// 	uiRoutes.GET("/email_failure", middleware.EnsureNotLoggedIn(), page.ShowEmailFailurePage)
		// 	uiRoutes.GET("/reset_password/:reset_password", middleware.EnsureNotLoggedIn(), page.ShowResetPasswordPage)

		uiRoutes.GET("/dashboard", page.DashboardPage)

		// 	uiRoutes.GET("/projects", middleware.EnsureLoggedIn(), page.ProjectsPageEndpoint)
		// 	uiRoutes.GET("/projects/:name", middleware.EnsureLoggedIn(), page.ProjectsPageEndpoint)

		// 	uiRoutes.GET("/runbooks", middleware.EnsureLoggedIn(), page.RunbooksPageEndpoint)
		// 	uiRoutes.GET("/runbooks/:runbook_id", middleware.EnsureLoggedIn(), api.RunbookKernelSetup)
		// 	uiRoutes.GET("/runbooks/:runbook_id/:kernel_id", middleware.EnsureLoggedIn(), page.RunbookPageEndpoint)

		// 	uiRoutes.GET("/alerts", middleware.EnsureLoggedIn(), page.AlertsPageEndpoint)
		// 	uiRoutes.GET("/alerts/:id", middleware.EnsureLoggedIn(), page.AlertPageEndpoint)

		// 	uiRoutes.GET("/runs", middleware.EnsureLoggedIn(), page.HistoriesPageEndpoint)
		// 	uiRoutes.GET("/runs/:run_id", middleware.EnsureLoggedIn(), page.HistoryPageEndpoint)

		// 	uiRoutes.GET("/users", middleware.EnsureLoggedIn(), page.UsersPageEndpoint)
		// 	uiRoutes.GET("/users/:username", middleware.EnsureLoggedIn(), page.UserPageEndpoint)

		// 	uiRoutes.GET("/401", common.Code401Endpoint)
		// 	uiRoutes.GET("/403", common.Code403Endpoint)
		// 	uiRoutes.GET("/404", common.Code404Endpoint)
		// 	uiRoutes.GET("/500", common.Code500Endpoint)
	}

	// htmxRoutes := router.Group("/htmx", middleware.CORSMiddleware(), middleware.EnsureLoggedInAPI())
	// {
	// 	// commonRoutes := htmxRoutes.Group("/common")
	// 	// {
	// 	// 	if err := common.Init(); err != nil {
	// 	// 		panic(err)
	// 	// 	}
	// 	// 	commonRoutes.GET("/status", common.StatusEndpoint)
	// 	// 	commonRoutes.GET("/sidebar", common.SidebarEndpoint)
	// 	// 	commonRoutes.GET("/error", common.ErrorEndpoint)
	// 	// 	commonRoutes.GET("/success", common.ErrorEndpoint)
	// 	// 	commonRoutes.GET("/header", common.HeaderEndpoint)
	// 	// }
	// 	projectRoutes := htmxRoutes.Group("/workflow")
	// 	{
	// 		projectRoutes.GET("/:name/display/:task", page.WorkflowDisplayEndpoint)
	// 		projectRoutes.GET("/:name/output/:task", page.WorkflowOutputEndpoint)
	// 		projectRoutes.GET("/:name/started/:task", page.WorkflowStartedEndpoint)
	// 		projectRoutes.GET("/:name/finished/:task", page.WorkflowFinishedEndpoint)
	// 		projectRoutes.GET("/:name/status/:task", page.WorkflowStatusEndpoint)
	// 		projectRoutes.GET("/:name/modal/:task", page.WorkflowModalEndpoint)
	// 	}
	// 	workflowsRoutes := htmxRoutes.Group("/workflows")
	// 	{
	// 		workflowsRoutes.GET("/table", page.WorkflowsTableEndpoint)
	// 		workflowsRoutes.GET("/search", page.WorkflowsSearchEndpoint)
	// 	}
	// 	dashboardRoutes := htmxRoutes.Group("/dashboard")
	// 	{
	// 		dashboardRoutes.GET("/table", page.DashboardTableEndpoint)
	// 		dashboardRoutes.GET("/search", page.DashboardSearchEndpoint)
	// 	}
	// 	runbooksRoutes := htmxRoutes.Group("/runbooks")
	// 	{
	// 		runbooksRoutes.GET("/table", page.RunbooksTableEndpoint)
	// 		runbooksRoutes.GET("/search", page.RunbooksSearchEndpoint)
	// 	}
	// 	alertsRoutes := htmxRoutes.Group("/alerts")
	// 	{
	// 		alertsRoutes.GET("/table", page.AlertsTableEndpoint)
	// 		alertsRoutes.GET("/search", page.AlertsSearchEndpoint)
	// 	}
	// 	alertRoutes := htmxRoutes.Group("/alert")
	// 	{
	// 		alertRoutes.POST("/:id/execute", page.AlertExecute)
	// 		alertRoutes.GET("/:alert_id/output/:kernel_id", page.AlertBuildOutput)
	// 	}
	// 	monitorsRoutes := htmxRoutes.Group("/monitors")
	// 	{
	// 		monitorsRoutes.GET("/table", page.MonitorsTableEndpoint)
	// 		monitorsRoutes.GET("/search", page.MonitorsSearchEndpoint)
	// 	}
	// 	monitorRoutes := htmxRoutes.Group("/monitor")
	// 	{
	// 		monitorRoutes.GET("/metadata/:id", page.MonitorMetadataEndpoint)
	// 		monitorRoutes.GET("/requirements/:id", page.MonitorRequirementsEndpoint)
	// 		monitorRoutes.GET("/contents/:id", page.MonitorContentsEndpoint)
	// 		monitorRoutes.GET("/enabled/:id", page.MonitorEnabledEndpoint)
	// 		monitorRoutes.GET("/status/:id", page.MonitorStatusEndpoint)
	// 		monitorRoutes.GET("/alerts/:id", page.MonitorAlertsEndpoint)
	// 	}
	// 	runsRoutes := htmxRoutes.Group("/runs")
	// 	{
	// 		runsRoutes.GET("/table", page.HistoriesTableEndpoint)
	// 		runsRoutes.GET("/search", page.HistoriesSearchEndpoint)
	// 		runsRoutes.GET("/timeline/:run_id", page.HistoryTimelineEndpoint)
	// 		runsRoutes.GET("/timeline/:run_id/status/:state_name", page.HistoryStateEndpoint)
	// 	}
	// 	usersRoutes := htmxRoutes.Group("/users")
	// 	{
	// 		usersRoutes.GET("/table", page.UsersTableEndpoint)
	// 		usersRoutes.GET("/search", page.UsersSearchEndpoint)
	// 	}
	// 	kernelRoutes := htmxRoutes.Group("/kernel")
	// 	{
	// 		kernelRoutes.POST("/execute", page.KernelExecute)
	// 		kernelRoutes.GET("/:kernel_id/:run_id/:block_idx", page.KernelBuildOutput)
	// 	}
	// }
}
