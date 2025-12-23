package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shelfarr/shelfarr/internal/auth"
	"github.com/shelfarr/shelfarr/internal/config"
	"github.com/shelfarr/shelfarr/internal/realtime"
	"gorm.io/gorm"
)

// Server represents the API server
type Server struct {
	config      *config.Config
	db          *gorm.DB
	echo        *echo.Echo
	authService *auth.AuthService
	wsHub       *realtime.Hub
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	e := echo.New()
	e.HideBanner = true

	// Create auth service
	authService := auth.NewAuthService(db, cfg.JWTSecret, 7*24*time.Hour)

	// Ensure admin user exists
	authService.EnsureAdminExists()

	// Create WebSocket hub
	wsHub := realtime.NewHub()
	go wsHub.Run()

	// Global Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	s := &Server{
		config:      cfg,
		db:          db,
		echo:        e,
		authService: authService,
		wsHub:       wsHub,
	}

	s.setupRoutes()

	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check (public)
	s.echo.GET("/health", s.healthCheck)

	// WebSocket endpoint (authenticated)
	s.echo.GET("/ws", s.wsHub.WebSocketHandler)

	// Auth handlers
	authHandlers := NewAuthHandlers(s.authService)

	// API v1 group
	api := s.echo.Group("/api/v1")

	// Public routes (no auth required)
	api.POST("/auth/login", authHandlers.Login)

	// Apply authentication middleware to protected routes
	// In development, auth can be optional. In production, enable it.
	protected := api.Group("")

	// SSO middleware (if enabled)
	if s.config.EnableSSO {
		protected.Use(auth.SSOMiddleware(s.authService, s.config.SSOHeaderName))
	}

	// JWT middleware for remaining requests
	protected.Use(auth.JWTMiddleware(s.authService))

	// Auth routes (protected)
	protected.POST("/auth/refresh", authHandlers.Refresh)
	protected.GET("/auth/me", authHandlers.GetCurrentUser)
	protected.POST("/auth/password", authHandlers.ChangePassword)

	// Library endpoints
	protected.GET("/library", s.getLibrary)
	protected.GET("/library/stats", s.getLibraryStats)

	// Book endpoints
	protected.GET("/books", s.getBooks)
	protected.GET("/books/:id", s.getBook)
	protected.POST("/books", s.addBook)
	protected.PUT("/books/bulk", s.bulkUpdateBooks)    // Must be before :id routes
	protected.DELETE("/books/bulk", s.bulkDeleteBooks) // Must be before :id routes
	protected.PUT("/books/:id", s.updateBook)
	protected.DELETE("/books/:id", s.deleteBook)
	protected.POST("/books/:bookId/search", s.automaticSearch)

	// Author endpoints
	protected.GET("/authors", s.getAuthors)
	protected.GET("/authors/:id", s.getAuthor)
	protected.POST("/authors", s.addAuthor)
	protected.PUT("/authors/:id", s.updateAuthor)
	protected.DELETE("/authors/:id", s.deleteAuthor)

	// Series endpoints
	protected.GET("/series", s.getSeries)
	protected.GET("/series/:id", s.getSeriesDetail)
	protected.POST("/series/:id/books", s.addSeriesBooks)

	// Search endpoints
	// Register POST route before GET to avoid path conflicts
	protected.POST("/search/hardcover/test", s.testHardcover)
	protected.GET("/search/hardcover", s.searchHardcover)
	protected.GET("/search/indexers", s.searchIndexers)

	// Hardcover detail endpoints (for viewing before adding)
	protected.GET("/hardcover/book/:id", s.getHardcoverBook)
	protected.GET("/hardcover/author/:id", s.getHardcoverAuthor)
	protected.GET("/hardcover/series/:id", s.getHardcoverSeries)
	protected.POST("/hardcover/book/:id", s.addHardcoverBook)

	// Indexer endpoints
	protected.GET("/indexers", s.getIndexers)
	protected.POST("/indexers", s.addIndexer)
	protected.PUT("/indexers/:id", s.updateIndexer)
	protected.DELETE("/indexers/:id", s.deleteIndexer)
	protected.POST("/indexers/:id/test", s.testIndexer)

	// Download client endpoints
	protected.GET("/downloadclients", s.getDownloadClients)
	protected.POST("/downloadclients", s.addDownloadClient)
	protected.POST("/downloadclients/test", s.testDownloadClientConfig) // Must be before :id routes
	protected.PUT("/downloadclients/:id", s.updateDownloadClient)
	protected.DELETE("/downloadclients/:id", s.deleteDownloadClient)
	protected.POST("/downloadclients/:id/test", s.testDownloadClient)

	// Media file endpoints
	protected.GET("/mediafiles", s.getMediaFiles)
	protected.GET("/mediafiles/:id/stream", s.streamMediaFile)
	protected.POST("/mediafiles/:id/kindle", s.sendToKindle)
	protected.DELETE("/mediafiles/:id", s.deleteMediaFile)

	// Import endpoints
	protected.GET("/import/pending", s.getPendingImports)
	protected.POST("/import/manual", s.manualImport)

	// Download endpoints
	protected.GET("/downloads", s.getDownloads)
	protected.GET("/downloads/:id", s.getDownload)
	protected.POST("/downloads", s.triggerDownload)
	protected.DELETE("/downloads/:id", s.deleteDownload)

	// User endpoints (admin only for some)
	protected.GET("/users", s.getUsers)
	protected.POST("/users", s.createUser)
	protected.GET("/users/me", s.getCurrentUser)
	protected.PUT("/users/:id", s.updateUser)

	// Progress tracking
	protected.GET("/progress/:mediaFileId", s.getProgress)
	protected.PUT("/progress/:mediaFileId", s.updateProgress)

	// Settings endpoints
	protected.GET("/settings", s.getSettings)
	protected.PUT("/settings", s.updateSettings)

	// General settings
	protected.GET("/settings/general", s.getGeneralSettings)
	protected.PUT("/settings/general", s.updateGeneralSettings)
	protected.GET("/settings/languages", s.getAvailableLanguages)

	// Media management settings
	protected.GET("/settings/media", s.getMediaSettings)
	protected.PUT("/settings/media", s.updateMediaSettings)
	protected.GET("/settings/media/naming-preview", s.getNamingPreview)

	// Filesystem browsing for directory selection
	protected.GET("/filesystem/browse", s.browseFilesystem)

	// Root folder endpoints
	protected.GET("/rootfolders", s.getRootFolders)
	protected.POST("/rootfolders", s.addRootFolder)
	protected.DELETE("/rootfolders/:id", s.deleteRootFolder)

	// Quality profile endpoints
	protected.GET("/profiles", s.getProfiles)
	protected.GET("/profiles/:id", s.getProfile)
	protected.POST("/profiles", s.createProfile)
	protected.PUT("/profiles/:id", s.updateProfile)
	protected.DELETE("/profiles/:id", s.deleteProfile)

	// Activity/History endpoint
	protected.GET("/activity", s.getActivity)
	protected.GET("/activity/history", s.getActivityHistory)

	// Wanted endpoint
	protected.GET("/wanted", s.getWanted)
	protected.GET("/wanted/missing", s.getWantedMissing)
	protected.GET("/wanted/cutoff", s.getWantedCutoff)

	// System endpoints
	protected.GET("/system/status", s.getSystemStatus)
	protected.GET("/system/tasks", s.getSystemTasks)
	protected.POST("/system/tasks/:name/run", s.runSystemTask)
	protected.GET("/system/logs", s.getSystemLogs)
	protected.POST("/system/backup", s.createBackup)

	// Notification endpoints
	protected.GET("/notifications", s.getNotifications)
	protected.POST("/notifications", s.addNotification)
	protected.PUT("/notifications/:id", s.updateNotification)
	protected.DELETE("/notifications/:id", s.deleteNotification)
	protected.POST("/notifications/:id/test", s.testNotification)

	// Hardcover List endpoints
	protected.GET("/lists", s.getLists)
	protected.POST("/lists", s.addList)
	protected.PUT("/lists/:id", s.updateList)
	protected.DELETE("/lists/:id", s.deleteList)
	protected.POST("/lists/:id/sync", s.syncList)

	// Serve static frontend files in production
	s.echo.Static("/", "public")
}

// Start begins listening for requests
func (s *Server) Start() error {
	return s.echo.Start(s.config.ListenAddr)
}

// GetWSHub returns the WebSocket hub for external use
func (s *Server) GetWSHub() *realtime.Hub {
	return s.wsHub
}

// healthCheck returns server health status
func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
	})
}
