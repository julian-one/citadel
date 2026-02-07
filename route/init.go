package route

import (
	"log/slog"
	"net/http"

	"citadel/internal/middleware"

	"github.com/jmoiron/sqlx"
)

type Config struct {
	Db     *sqlx.DB
	Logger *slog.Logger
}

func Initialize(config Config) http.Handler {
	baseChain := middleware.New(
		middleware.Logger,
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

	// -----------------
	// Public routes
	// -----------------
	mux.Handle("GET /health", baseChain.ThenFunc(GetHealth()))
	mux.Handle("POST /register", baseChain.ThenFunc(Register(config.Logger, config.Db)))
	mux.Handle("POST /login", baseChain.ThenFunc(Login(config.Logger, config.Db)))
	mux.Handle("POST /logout", baseChain.ThenFunc(Logout(config.Logger, config.Db)))
	mux.Handle("GET /session/{id}", baseChain.ThenFunc(GetSession(config.Db)))

	// -----------------
	// Protected routes
	// -----------------
	mux.Handle("GET /user", protectedChain.ThenFunc(ListUsers(config.Db)))
	mux.Handle("GET /user/{id}", protectedChain.ThenFunc(GetUser(config.Logger, config.Db)))
	mux.Handle("PATCH /user/{id}", protectedChain.ThenFunc(UpdateUser(config.Logger, config.Db)))
	mux.Handle("GET /pokemon", protectedChain.ThenFunc(SearchPokemon(config.Logger, config.Db)))

	return mux
}
