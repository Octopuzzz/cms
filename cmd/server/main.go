// CMS Backend Server
//
// @title           CMS Backend API
// @version         1.0.0
// @description     High-performance, scalable CMS Backend with dynamic service creation,
//
//	multi-database support, role-based access control, and microservices integration.
//
// @contact.name   API Support
// @contact.email  support@cms-backend.com
//
// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT
//
// @host           localhost:8080
// @BasePath       /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in             header
// @name           Authorization
// @description    Type "Bearer" followed by a space and the JWT token.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cms-backend/internal/config"
	"cms-backend/internal/database"
	"cms-backend/internal/handlers"
	"cms-backend/internal/middleware"
	"cms-backend/internal/services"
	"cms-backend/internal/tracing"
	"cms-backend/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var Version = "1.0.0"

func main() {
	// Logger
	log := logger.GetLogger()

	// Config
	cfg := config.GetInstance()
	logger.InitLogger(cfg.App.LogLevel, "json")

	log.Info("Starting CMS Backend",
		"name", cfg.App.Name,
		"version", Version,
		"env", cfg.App.Environment,
	)

	// Tracing
	tp, err := tracing.InitTracer(&cfg.Jaeger)
	if err != nil {
		log.Warn("Tracing init failed", "error", err)
	}
	defer tracing.Shutdown(context.Background(), tp)

	// Database
	connManager := database.GetConnectionManager()
	if err := connManager.Initialize(cfg); err != nil {
		log.Fatal("Database init failed", "error", err)
		os.Exit(1)
	}
	defer connManager.Close()

	defaultConn, err := connManager.GetDefaultConnection()
	if err != nil {
		log.Fatal("No default DB connection", "error", err)
		os.Exit(1)
	}
	db := defaultConn.DB

	// Services
	authSvc := services.NewAuthService(db, &cfg.JWT)
	svcService := services.NewServiceService(connManager)
	dynamicDataSvc := services.NewDynamicDataService(connManager, svcService)
	graphqlSvc := services.NewGraphQLService(dynamicDataSvc, svcService)
	userSvc := services.NewUserService(db)
	roleSvc := services.NewRoleService(db)
	dbConnSvc := services.NewDBConnService(db)
	valSvc := services.NewValidationService(db)
	menuSvc := services.NewMenuService(db)
	migSvc := services.NewMigrationService(connManager)
	backupSvc := services.NewBackupService(connManager)

	// Handlers
	authH := handlers.NewAuthHandler(authSvc)
	svcH := handlers.NewServiceHandler(svcService)
	dataH := handlers.NewDynamicDataHandler(dynamicDataSvc, svcService)
	graphqlH := handlers.NewGraphQLHandler(graphqlSvc)
	userH := handlers.NewUserHandler(userSvc)
	roleH := handlers.NewRoleHandler(roleSvc)
	dbConnH := handlers.NewDBConnectionHandler(dbConnSvc, connManager)
	menuH := handlers.NewMenuHandler(menuSvc)
	valH := handlers.NewValidationHandler(valSvc)
	migH := handlers.NewMigrationHandler(migSvc)
	backupH := handlers.NewBackupHandler(backupSvc)

	// Router
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.GinMiddleware())
	router.Use(middleware.SecurityHeaders())

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Correlation-ID", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Correlation-ID", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rate limiter
	if cfg.RateLimit.Enabled {
		router.Use(middleware.RateLimiter(cfg.RateLimit.RPS, cfg.RateLimit.Burst))
	}

	// Prometheus metrics middleware
	router.Use(middleware.PrometheusMetrics())

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "version": Version, "timestamp": time.Now().UTC()})
	})
	router.GET("/ready", func(c *gin.Context) {
		if _, err := connManager.GetDefaultConnection(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	setupRoutes(v1, authSvc, authH, svcH, dataH, graphqlH, userH, roleH, dbConnH, menuH, valH, migH, backupH)

	// HTTP Server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Info("HTTP server listening", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Shutdown error", "error", err)
	}
	log.Info("Server stopped")
}

func setupRoutes(
	v1 *gin.RouterGroup,
	authSvc *services.AuthService,
	authH *handlers.AuthHandler,
	svcH *handlers.ServiceHandler,
	dataH *handlers.DynamicDataHandler,
	graphqlH *handlers.GraphQLHandler,
	userH *handlers.UserHandler,
	roleH *handlers.RoleHandler,
	dbConnH *handlers.DBConnectionHandler,
	menuH *handlers.MenuHandler,
	valH *handlers.ValidationHandler,
	migH *handlers.MigrationHandler,
	backupH *handlers.BackupHandler,
) {
	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authH.Login)
		auth.POST("/register", authH.Register)
		auth.POST("/refresh", authH.Refresh)
	}

	// Authenticated routes
	protected := v1.Group("")
	protected.Use(middleware.Auth(authSvc))
	{
		protected.GET("/auth/me", authH.Me)

		// Menu (read accessible to all authenticated users)
		protected.GET("/menu", menuH.GetMenuTree)

		// DB Connections (admin only)
		dbConn := protected.Group("/database-connections")
		dbConn.Use(middleware.RequireRole("admin", "super_admin"))
		{
			dbConn.GET("", dbConnH.ListConnections)
			dbConn.POST("", dbConnH.CreateConnection)
			dbConn.GET("/:id", dbConnH.GetConnection)
			dbConn.DELETE("/:id", dbConnH.DeleteConnection)
			dbConn.POST("/test", dbConnH.TestConnection)
		}

		// Services (admin only create/update/delete, all can read)
		svcs := protected.Group("/services")
		{
			svcs.GET("", svcH.ListServices)
			svcs.GET("/:id", svcH.GetService)
			svcs.POST("", middleware.RequireRole("admin", "super_admin"), svcH.CreateService)
			svcs.PUT("/:id", middleware.RequireRole("admin", "super_admin"), svcH.UpdateService)
			svcs.DELETE("/:id", middleware.RequireRole("admin", "super_admin"), svcH.DeleteService)
			svcs.POST("/:id/permissions", middleware.RequireRole("admin", "super_admin"), svcH.SetPermissions)
		}

		// Dynamic data (RBAC handled inside handler)
		data := protected.Group("/data")
		{
			data.GET("/:slug", dataH.ListData)
			data.POST("/:slug", dataH.CreateData)
			data.GET("/:slug/:id", dataH.GetDataByID)
			data.PUT("/:slug/:id", dataH.UpdateData)
			data.DELETE("/:slug/:id", dataH.DeleteData)

			// Optional GraphQL Gateway route
			data.POST("/:slug/graphql", graphqlH.HandleGraphQL)
		}

		// Users (admin only)
		users := protected.Group("/users")
		users.Use(middleware.RequireRole("admin", "super_admin"))
		{
			users.GET("", userH.ListUsers)
			users.POST("", userH.CreateUser)
			users.GET("/:id", userH.GetUser)
			users.PUT("/:id", userH.UpdateUser)
			users.DELETE("/:id", userH.DeleteUser)
		}
		// Self password change
		protected.PUT("/users/me/password", userH.ChangePassword)

		// Roles
		roles := protected.Group("/roles")
		{
			roles.GET("", roleH.ListRoles)
			roles.POST("", middleware.RequireRole("admin", "super_admin"), roleH.CreateRole)
			roles.GET("/:id", roleH.GetRole)
			roles.PUT("/:id", middleware.RequireRole("admin", "super_admin"), roleH.UpdateRole)
			roles.DELETE("/:id", middleware.RequireRole("admin", "super_admin"), roleH.DeleteRole)
		}
		protected.GET("/permissions", roleH.ListPermissions)

		// Custom Validations
		vals := protected.Group("/validations")
		{
			vals.GET("", valH.ListValidations)
			vals.POST("", middleware.RequireRole("admin", "super_admin"), valH.CreateValidation)
		}

		// CMS Control Plane
		cms := protected.Group("/cms")
		cms.Use(middleware.RequireRole("admin", "super_admin"))
		{
			// Database connections (CMS aliases)
			cms.GET("/databases", dbConnH.ListConnections)
			cms.POST("/databases", dbConnH.CreateConnection)

			// Services (CMS aliases)
			cms.GET("/services", svcH.ListServices)
			cms.POST("/services", svcH.CreateService)

			// Relations are managed via service field relation_config

			// Migrations
			cms.POST("/migrations", migH.CreateMigration)
			cms.GET("/migrations/service/:service_id", migH.ListMigrations)
			cms.POST("/migrations/:id/rollback", migH.RollbackMigration)

			// Backups
			cms.POST("/backup/service/:service_id", backupH.CreateBackup)
			cms.GET("/backup/service/:service_id", backupH.ListBackups)
			cms.POST("/restore/:id", backupH.RestoreBackup)
		}
	}
}
