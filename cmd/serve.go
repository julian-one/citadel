package cmd

import (
	"log/slog"
	"net/http"
	"os"

	"citadel/internal/database"
	"citadel/internal/logging"
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

	// Initialize route handlers
	handler := route.Initialize(route.Config{
		Db:     db,
		Logger: logger,
	})

	// Start HTTP server
	port := viper.GetString("server.port")
	logger.InfoContext(ctx, "Server listening", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logger.ErrorContext(ctx, "failed to start server", "error", err)
		os.Exit(1)
	}
}
