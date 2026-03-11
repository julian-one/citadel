package cmd

import (
	"fmt"
	"log/slog"
	"net/http"

	"citadel/internal/database"
	"citadel/internal/logging"
	"citadel/internal/parser"
	"citadel/route"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:          "serve",
	Short:        "Start the Citadel HTTP server",
	Long:         `The serve command starts the Citadel HTTP web server`,
	SilenceUsage: true,
	RunE:         runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringP("port", "p", "8080", "port to listen on")
	serveCmd.Flags().String("db-path", "./citadel.db", "path to the SQLite database")
	serveCmd.Flags().String("db-schema", "./schema/model.sql", "path to the database schema file")

	_ = viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
	_ = viper.BindPFlag("database.path", serveCmd.Flags().Lookup("db-path"))
	_ = viper.BindPFlag("database.schema", serveCmd.Flags().Lookup("db-schema"))
}

func runServe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Initialize logger
	logger := logging.New(slog.LevelInfo)
	slog.SetDefault(logger)

	// Initialize database
	db, err := database.New(viper.GetString("database.path"), viper.GetString("database.schema"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
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
		return fmt.Errorf("server stopped: %w", err)
	}

	return nil
}
