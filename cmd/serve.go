package cmd

import (
	"log/slog"
	"net/http"
	"os"

	"citadel/internal/database"
	"citadel/internal/logging"
	"citadel/internal/parser"
	"citadel/route"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Citadel HTTP server",
	Long:  `The serve command starts the Citadel HTTP web server`,
	Run:   runServe,
}

func runServe(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.path", "./citadel.db")
	viper.SetDefault("database.schema", "./schema/model.sql")
	viper.SetDefault("server.max_upload_mb", 10)
	viper.SetDefault("anthropic.model", "claude-sonnet-4-5-20250929")
	viper.SetDefault("anthropic.api_key", "")

	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		slog.ErrorContext(ctx, "failed to read config file", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := logging.New(slog.LevelInfo)
	slog.SetDefault(logger)

	// Initialize database
	db, err := database.New(viper.GetString("database.path"), viper.GetString("database.schema"))
	if err != nil {
		logger.ErrorContext(ctx, "failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize parser
	claude := parser.New(viper.GetString("anthropic.api_key"), viper.GetString("anthropic.model"))

	// Initialize route handlers
	handler := route.Initialize(route.Config{
		Logger: logger,
		Db:     db,
		Parser: claude,
	})

	// Start HTTP server
	port := viper.GetString("server.port")
	logger.InfoContext(ctx, "Server listening", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logger.ErrorContext(ctx, "failed to start server", "error", err)
		os.Exit(1)
	}
}
