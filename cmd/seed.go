package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"citadel/internal/database"
	"citadel/internal/logging"
	"citadel/internal/session"
	"citadel/internal/user"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with a test admin user",
	Long:  `Creates an admin user (username: admin, password: password1234) and prints a session ID for testing.`,
	RunE:  runSeed,
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

func runSeed(cmd *cobra.Command, args []string) error {
	logger := logging.New(slog.LevelInfo)
	slog.SetDefault(logger)

	db, err := database.New(
		viper.GetString("database.path"),
		viper.GetString("database.schema"),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	uid, err := user.Create(ctx, db, user.CreateRequest{
		Username: "admin",
		Email:    "admin@test.com",
		Password: "password1234",
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	_, err = db.ExecContext(
		ctx,
		`UPDATE users SET role = 'admin' WHERE user_id = ?`,
		uid,
	)
	if err != nil {
		return fmt.Errorf("failed to promote user to admin: %w", err)
	}

	sess, err := session.Create(ctx, db, uid)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Printf("admin user created (id: %s)\n", uid)
	fmt.Printf("session_id: %s\n", sess.SessionId)

	return nil
}
