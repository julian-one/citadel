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
	Long:  `The serve command starts the Citadel HTTP server using the configuration`,
	Run:   runServe,
}

func runServe(cmd *cobra.Command, args []string) {
	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.path", "./citadel.db")
	viper.SetDefault("database.schema", "./schema/model.sql")

	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("Failed to read config file", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := logging.New(slog.LevelInfo)
	slog.SetDefault(logger)

	logger.Info("Starting citadel")

	// Initialize database
	db, err := database.New(viper.GetString("database.path"), viper.GetString("database.schema"))
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize routes
	config := route.Config{
		Db:     db,
		Logger: logger,
	}
	handler := route.Initialize(config)

	// Start server
	port := viper.GetString("server.port")
	logger.Info("Server listening", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
