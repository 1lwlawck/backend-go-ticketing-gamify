package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-go-ticketing-gamify/internal/audit"
	"backend-go-ticketing-gamify/internal/auth"
	"backend-go-ticketing-gamify/internal/config"
	"backend-go-ticketing-gamify/internal/epics"
	"backend-go-ticketing-gamify/internal/gamification"
	"backend-go-ticketing-gamify/internal/middleware"
	"backend-go-ticketing-gamify/internal/projects"
	"backend-go-ticketing-gamify/internal/tickets"
	"backend-go-ticketing-gamify/internal/users"
)

const serviceVersion = "0.1.0"

// Server exposes HTTP endpoints for the ticketing service.
type Server struct {
	cfg  config.Config
	pool *pgxpool.Pool
}

// New builds a Server with the provided Config and db pool.
func New(cfg config.Config, pool *pgxpool.Pool) *Server {
	return &Server{cfg: cfg, pool: pool}
}

// Start runs the HTTP server until context is canceled.
func (s *Server) Start(ctx context.Context) error {
	engine := s.routes()
	srv := &http.Server{
		Addr:              s.cfg.Addr(),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("HTTP server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) routes() *gin.Engine {
	if s.cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.CORS(),
		middleware.RateLimit(s.cfg.RateLimitPerMin, s.cfg.APIKeyRateLimit, s.cfg.RateLimitWindow, s.cfg.APIKeyHeader),
	)
	if s.cfg.Env != "production" {
		engine.Use(gin.Logger())
	}

	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "ticketing-gamify",
			"status":  "ok",
		})
	})

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"health": "ok"})
	})

	engine.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": serviceVersion})
	})

	api := engine.Group("/api/v1")
	api.Use(middleware.APIKeyGuard(s.cfg.APIKey, s.cfg.APIKeyHeader))

	auditRepo := audit.NewRepository(s.pool)
	auditSvc := audit.NewService(auditRepo)
	auditHandler := audit.NewHandler(auditSvc)

	gamRepo := gamification.NewRepository(s.pool)
	gamSvc := gamification.NewService(gamRepo)
	gamHandler := gamification.NewHandler(gamSvc)

	authRepo := auth.NewRepository(s.pool)
	authSvc := auth.NewService(authRepo, gamSvc, auditSvc, s.cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc)
	authHandler.RegisterRoutes(api.Group("/auth"))

	userRepo := users.NewRepository(s.pool)
	userSvc := users.NewService(userRepo, auditSvc)
	userHandler := users.NewHandler(userSvc)

	projectRepo := projects.NewRepository(s.pool)
	projectSvc := projects.NewService(projectRepo, auditSvc)
	projectHandler := projects.NewHandler(projectSvc)

	ticketRepo := tickets.NewRepository(s.pool)
	ticketSvc := tickets.NewService(ticketRepo, auditSvc, gamSvc)
	ticketHandler := tickets.NewHandler(ticketSvc)

	epicRepo := epics.NewRepository(s.pool)
	epicSvc := epics.NewService(epicRepo)
	epicHandler := epics.NewHandler(epicSvc)

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))

	projectHandler.RegisterRoutes(protected.Group("/projects"))
	epicHandler.RegisterRoutes(protected)
	ticketHandler.RegisterRoutes(protected.Group("/tickets"))
	gamHandler.RegisterRoutes(protected.Group("/gamification"))

	authProtected := protected.Group("/auth")
	authHandler.RegisterProtected(authProtected)

	usersGroup := protected.Group("/users")
	userHandler.RegisterRoutes(usersGroup)

	auditGroup := protected.Group("/audit")
	auditGroup.Use(middleware.RequireRoles("admin", "project_manager"))
	auditHandler.RegisterRoutes(auditGroup)

	return engine
}
