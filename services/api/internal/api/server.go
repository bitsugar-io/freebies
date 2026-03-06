package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
	authmw "github.com/retr0h/freebie/services/api/internal/api/middleware"
	workergen "github.com/retr0h/freebie/services/api/internal/api/worker/gen"
	"github.com/retr0h/freebie/services/api/internal/config"
	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/worker"
)

type Server struct {
	cfg           *config.Config
	db            *sql.DB
	router        *chi.Mux
	logger        *slog.Logger
	workerService *worker.Service
}

func NewServer(cfg *config.Config, db *sql.DB, logger *slog.Logger, workerService *worker.Service) *Server {
	s := &Server{
		cfg:           cfg,
		db:            db,
		router:        chi.NewRouter(),
		logger:        logger,
		workerService: workerService,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(30 * time.Second))

	// CORS for mobile/web app
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all for dev, restrict in prod
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Device-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health checks
	s.router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// API routes
	h := handlers.New(s.db, s.logger)
	queries := db.New(s.db)

	s.router.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		r.Get("/leagues", h.ListLeagues)
		r.Get("/events", h.ListEvents)
		r.Get("/events/{id}", h.GetEvent)
		r.Post("/users", h.CreateUser) // Returns token on creation

		// Protected routes (require auth + user must match)
		r.Group(func(r chi.Router) {
			r.Use(authmw.Auth(queries))
			r.Use(authmw.RequireUserMatch)

			// User profile
			r.Get("/users/{id}", h.GetUser)
			r.Get("/users/{id}/stats", h.GetUserStats)
			r.Put("/users/{id}/push-token", h.UpdatePushToken)

			// Subscriptions
			r.Get("/users/{userId}/subscriptions", h.ListSubscriptions)
			r.Post("/users/{userId}/subscriptions", h.CreateSubscription)
			r.Delete("/users/{userId}/subscriptions/{eventId}", h.DeleteSubscription)

			// Active Deals
			r.Get("/users/{userId}/active-deals", h.ListActiveDeals)

			// Dismissals
			r.Post("/users/{userId}/dismissals", h.CreateDismissal)
			r.Delete("/users/{userId}/dismissals/{triggeredEventId}", h.DeleteDismissal)
		})
	})

	// Internal worker endpoints (called by K8s CronJobs via generated client)
	if s.workerService != nil {
		wh := handlers.NewWorkerHandler(s.workerService)
		strictHandler := workergen.NewStrictHandler(wh, nil)

		// Auth middleware applied to the generated routes
		internalRouter := chi.NewRouter()
		internalRouter.Use(authmw.InternalAuth(s.cfg.Worker.Secret))
		workergen.HandlerFromMux(strictHandler, internalRouter)
		s.router.Mount("/", internalRouter)
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) Router() http.Handler {
	return s.router
}
