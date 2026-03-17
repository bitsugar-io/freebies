package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/retr0h/freebie/services/api/internal/api"
	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
	"github.com/retr0h/freebie/services/api/internal/triggers"
	"github.com/retr0h/freebie/services/api/internal/worker"

	_ "github.com/retr0h/freebie/services/api/internal/sources/mlb"
	_ "github.com/retr0h/freebie/services/api/internal/sources/test"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long:  `Start the HTTP API server that serves events and handles subscriptions.`,
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().String("host", "0.0.0.0", "Host to bind to")
	serveCmd.Flags().Int("port", 8080, "Port to listen on")

	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
}

func runServe(cmd *cobra.Command, args []string) error {
	logger.Info("config loaded",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"database", cfg.Database.Path,
	)

	// Open database
	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		logger.Error("failed to open database", "error", err, "path", cfg.Database.Path)
		return err
	}
	defer database.Close()

	// Run migrations
	logger.Info("running migrations...")
	if err := db.Migrate(database); err != nil {
		logger.Error("failed to run migrations", "error", err)
		return err
	}
	logger.Info("database ready", "path", cfg.Database.Path)

	// Create worker service for internal endpoints
	queries := db.New(database)
	checker := triggers.NewChecker(queries)
	notifier := notify.NewExpoNotifier()
	workerService := worker.NewService(queries, checker, notifier, logger)

	// Context with signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start server
	server := api.NewServer(&cfg, database, logger, workerService)
	return server.Start(ctx)
}
