package cmd

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/retr0h/freebie/internal/config"
)

var (
	cfgFile    string
	debug      bool
	jsonOutput bool
	logger     = slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg        config.Config
)

var rootCmd = &cobra.Command{
	Use:   "freebie",
	Short: "Freebie - Sports rewards notification service",
	Long: `Freebie monitors sports events and sends notifications
when your favorite teams trigger special offers.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output logs in JSON format")
	rootCmd.PersistentFlags().String("db", "", "database file path")

	viper.BindPFlag("database.path", rootCmd.PersistentFlags().Lookup("db"))

	// Add subcommands
	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	viper.SetEnvPrefix("FREEBIE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Defaults (server.host/port are set via serve.go flags)
	viper.SetDefault("database.path", "freebie.db")

	// Config file is optional
	_ = viper.ReadInConfig()

	// Unmarshal into global config
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Error("failed to unmarshal config", "error", err)
		os.Exit(1)
	}
}

func initLogger() {
	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	if jsonOutput {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	} else {
		logger = slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      logLevel,
				TimeFormat: time.Kitchen,
				NoColor:    !term.IsTerminal(int(os.Stdout.Fd())),
			}),
		)
	}

	slog.SetDefault(logger)
}
