package route

import (
	"net/http"

	"citadel/internal/middleware"

	"github.com/jmoiron/sqlx"
)

type Config struct {
	Db *sqlx.DB
}

func Initialize(config Config) http.Handler {
	baseChain := middleware.New(
		middleware.CORS,
	)
	protectedChain := baseChain.Use(
		middleware.RequireAuth(config.Db),
	)

	mux := http.NewServeMux()

	// Handle OPTIONS preflight for all routes
	mux.Handle("OPTIONS /", baseChain.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Public routes
	mux.Handle("GET /health", baseChain.ThenFunc(GetHealth()))
	mux.Handle("POST /register", baseChain.ThenFunc(Register(config.Db)))
	mux.Handle("POST /login", baseChain.ThenFunc(Login(config.Db)))

	// Protected routes
	mux.Handle("GET /me", protectedChain.ThenFunc(GetMe()))
	mux.Handle("POST /logout", protectedChain.ThenFunc(Logout(config.Db)))
	mux.Handle("GET /users", protectedChain.ThenFunc(ListUsers(config.Db)))

	return mux
}
