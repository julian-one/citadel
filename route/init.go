package route

import (
	"log/slog"
	"net/http"

	"citadel/internal/email"
	"citadel/internal/middleware"
	"citadel/internal/parser"

	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
)

type Config struct {
	Logger     *slog.Logger
	Db         *sqlx.DB
	Parser     *parser.Claude
	Email      *email.Client
	SigningKey string
}

func Initialize(config Config) http.Handler {
	baseChain := middleware.New(
		middleware.Logger(config.Logger),
	)
	optionalChain := baseChain.Use(
		middleware.OptionalAuthentication(config.Db),
	)
	protectedChain := baseChain.Use(
		middleware.Authentication(config.Db),
	)
	// NOTE: Admin middleware is used as part of the protected chain
	adminChain := protectedChain.Use(
		middleware.Admin(config.Db),
	)

	mux := http.NewServeMux()

	// -----------------
	// Public
	// -----------------

	// Health
	mux.Handle("GET /health", baseChain.ThenFunc(GetHealth()))

	// Auth
	mux.Handle(
		"POST /register",
		baseChain.ThenFunc(Register(config.Logger, config.Db, config.Email, config.SigningKey)),
	)
	mux.Handle(
		"POST /register/verify",
		baseChain.ThenFunc(VerifyRegistration(config.Logger, config.SigningKey)),
	)
	mux.Handle(
		"POST /register/complete",
		baseChain.ThenFunc(CompleteRegistration(config.Logger, config.Db, config.SigningKey)),
	)
	mux.Handle("POST /login", baseChain.ThenFunc(Login(config.Logger, config.Db)))
	mux.Handle("POST /logout", baseChain.ThenFunc(Logout(config.Logger, config.Db)))

	// Sessions
	mux.Handle("GET /sessions/{id}", baseChain.ThenFunc(GetSession(config.Logger, config.Db)))

	// Recipes
	mux.Handle("GET /recipes", baseChain.ThenFunc(ListRecipes(config.Logger, config.Db)))
	mux.Handle("GET /recipes/{id}", baseChain.ThenFunc(GetRecipe(config.Logger, config.Db)))

	// -----------------
	// Optional
	// -----------------

	// Posts
	mux.Handle("GET /posts", optionalChain.ThenFunc(ListPosts(config.Logger, config.Db)))
	mux.Handle("GET /posts/{id}", optionalChain.ThenFunc(GetPost(config.Logger, config.Db)))

	// -----------------
	// Protected
	// -----------------

	// Users
	mux.Handle("GET /users/{id}", protectedChain.ThenFunc(GetUser(config.Logger, config.Db)))
	mux.Handle("PATCH /users/{id}", protectedChain.ThenFunc(UpdateUser(config.Logger, config.Db)))
	mux.Handle(
		"PATCH /users/{id}/password",
		protectedChain.ThenFunc(UpdatePassword(config.Logger, config.Db)),
	)

	// Posts
	mux.Handle("POST /posts", protectedChain.ThenFunc(CreatePost(config.Logger, config.Db)))
	mux.Handle("PATCH /posts/{id}", protectedChain.ThenFunc(UpdatePost(config.Logger, config.Db)))
	mux.Handle("DELETE /posts/{id}", protectedChain.ThenFunc(DeletePost(config.Logger, config.Db)))

	// Recipes
	mux.Handle("POST /recipes", protectedChain.ThenFunc(CreateRecipe(config.Logger, config.Db)))
	mux.Handle(
		"PATCH /recipes/{id}",
		protectedChain.ThenFunc(UpdateRecipe(config.Logger, config.Db)),
	)
	mux.Handle(
		"DELETE /recipes/{id}",
		protectedChain.ThenFunc(DeleteRecipe(config.Logger, config.Db)),
	)

	// Pokemon
	mux.Handle("GET /pokemon", protectedChain.ThenFunc(SearchPokemon(config.Logger, config.Db)))

	// -----------------
	// Admin
	// -----------------

	// Users
	mux.Handle("GET /users", adminChain.ThenFunc(ListUsers(config.Logger, config.Db)))
	mux.Handle(
		"PATCH /users/{id}/role",
		adminChain.ThenFunc(UpdateUserRole(config.Logger, config.Db)),
	)

	// Sessions
	mux.Handle(
		"GET /users/{id}/sessions",
		adminChain.ThenFunc(ListSessions(config.Logger, config.Db)),
	)
	mux.Handle(
		"DELETE /users/{id}/sessions",
		adminChain.ThenFunc(DeleteAllSessions(config.Logger, config.Db)),
	)
	mux.Handle(
		"DELETE /sessions/{id}",
		adminChain.ThenFunc(DeleteSession(config.Logger, config.Db)),
	)

	// Recipes
	mux.Handle(
		"POST /recipes/scan",
		adminChain.ThenFunc(ScanRecipe(config.Logger, config.Parser)),
	)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://julian-one.com", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Cache-Control"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	return c.Handler(mux)
}
