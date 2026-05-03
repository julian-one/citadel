package cmd

import (
	"fmt"
	"log/slog"
	"net/http"

	"citadel/internal/broker"
	"citadel/internal/database"
	"citadel/internal/email"
	"citadel/internal/logger"
	"citadel/internal/parser"
	"citadel/route"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Citadel HTTP server",
	Long:  `The serve command starts the Citadel HTTP web server`,
	RunE:  runServe,
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
	// Initialize logger
	l := logger.New(slog.LevelInfo)
	slog.SetDefault(l)

	// Initialize database
	db, err := database.New(viper.GetString("database.path"), viper.GetString("database.schema"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize parser
	claude := parser.New(viper.GetString("anthropic.api_key"), viper.GetString("anthropic.model"))

	// Initialize email client
	baseURL := viper.GetString("server.base_url")
	emailClient := email.New(
		viper.GetString("resend.api_key"),
		viper.GetString("resend.from_email"),
		baseURL,
	)

	signingKey := viper.GetString("hmac.signing_key")

	// Initialize route handlers
	handler := route.Initialize(route.Config{
		Logger:     l,
		DB:         db,
		Parser:     claude,
		Email:      emailClient,
		SigningKey: signingKey,
		Broker: broker.New(
			viper.GetString("alpaca.key"),
			viper.GetString("alpaca.secret"),
			viper.GetString("alpaca.endpoint"),
		),
	})

	// Start HTTP server
	port := viper.GetString("server.port")
	l.Info("server listening", "port", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		return fmt.Errorf("server stopped: %w", err)
	}

	return nil
}
