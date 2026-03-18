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

	"github.com/retr0h/freebie/services/scheduler/internal/config"
)

var (
	cfgFile    string
	debug      bool
	jsonOutput bool
	logger     = slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg        config.Config
)

var rootCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Freebie scheduled task runner",
	Long: `Scheduler calls the Freebie API's internal worker
endpoints on a schedule to check triggers and send reminders.`,

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

	// Add subcommands
	rootCmd.AddCommand(checkTriggersCmd)
	rootCmd.AddCommand(sendRemindersCmd)
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

	viper.SetEnvPrefix("SCHEDULER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("api.url", "http://freebie-api:8080")
	viper.SetDefault("worker.secret", "")

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
