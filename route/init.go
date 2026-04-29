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
	DB         *sqlx.DB
	Parser     *parser.Claude
	Email      *email.Client
	SigningKey string
}

func Initialize(config Config) http.Handler {
	baseChain := middleware.New(
		middleware.Logger(config.Logger),
	)
	optionalChain := baseChain.Append(
		middleware.OptionalAuthentication(config.DB),
	)
	protectedChain := baseChain.Append(
		middleware.Authentication(config.DB),
	)
	// NOTE: Admin middleware is used as part of the protected chain
	adminChain := protectedChain.Append(
		middleware.Admin(config.DB),
	)

	mux := http.NewServeMux()

	// -----------------
	// Health
	// -----------------
	mux.Handle("GET /health", baseChain.Wrap(Health()))

	// -----------------
	// Auth
	// -----------------
	mux.Handle(
		"POST /register",
		baseChain.Wrap(Register(config.Logger, config.DB, config.Email, config.SigningKey)),
	)
	mux.Handle(
		"POST /register/verify",
		baseChain.Wrap(VerifyRegistration(config.Logger, config.SigningKey)),
	)
	mux.Handle(
		"POST /register/complete",
		baseChain.Wrap(CompleteRegistration(config.Logger, config.DB, config.SigningKey)),
	)
	mux.Handle("POST /login", baseChain.Wrap(Login(config.Logger, config.DB)))
	mux.Handle("POST /logout", baseChain.Wrap(Logout(config.Logger, config.DB)))

	// -----------------
	// Users
	// -----------------
	mux.Handle("GET /users", adminChain.Wrap(ListUsers(config.Logger, config.DB)))
	mux.Handle("GET /users/{id}", protectedChain.Wrap(GetUser(config.Logger, config.DB)))
	mux.Handle("PATCH /users/{id}", protectedChain.Wrap(UpdateUser(config.Logger, config.DB)))
	mux.Handle(
		"PATCH /users/{id}/password",
		protectedChain.Wrap(UpdatePassword(config.Logger, config.DB)),
	)
	mux.Handle(
		"PATCH /users/{id}/role",
		adminChain.Wrap(UpdateUserRole(config.Logger, config.DB)),
	)

	// -----------------
	// Sessions
	// -----------------
	mux.Handle("GET /sessions/{id}", protectedChain.Wrap(GetSession(config.Logger, config.DB)))
	mux.Handle("DELETE /sessions/{id}", adminChain.Wrap(DeleteSession(config.Logger, config.DB)))
	mux.Handle(
		"GET /users/{id}/sessions",
		adminChain.Wrap(ListSessions(config.Logger, config.DB)),
	)
	mux.Handle(
		"DELETE /users/{id}/sessions",
		adminChain.Wrap(DeleteAllSessions(config.Logger, config.DB)),
	)

	// -----------------
	// Posts
	// -----------------
	mux.Handle("GET /posts", optionalChain.Wrap(ListPosts(config.Logger, config.DB)))
	mux.Handle("GET /posts/{id}", optionalChain.Wrap(GetPost(config.Logger, config.DB)))
	mux.Handle("POST /posts", protectedChain.Wrap(CreatePost(config.Logger, config.DB)))
	mux.Handle("PATCH /posts/{id}", protectedChain.Wrap(UpdatePost(config.Logger, config.DB)))
	mux.Handle("DELETE /posts/{id}", protectedChain.Wrap(DeletePost(config.Logger, config.DB)))

	// -----------------
	// Recipes
	// -----------------
	mux.Handle("GET /recipes", optionalChain.Wrap(ListRecipes(config.Logger, config.DB)))
	mux.Handle("GET /recipes/{id}", optionalChain.Wrap(GetRecipe(config.Logger, config.DB)))
	mux.Handle("POST /recipes", protectedChain.Wrap(CreateRecipe(config.Logger, config.DB)))
	mux.Handle(
		"PATCH /recipes/{id}",
		protectedChain.Wrap(UpdateRecipe(config.Logger, config.DB)),
	)
	mux.Handle(
		"DELETE /recipes/{id}",
		protectedChain.Wrap(DeleteRecipe(config.Logger, config.DB)),
	)
	mux.Handle(
		"POST /recipes/scan",
		adminChain.Wrap(ScanRecipe(config.Logger, config.Parser)),
	)

	// -----------------
	// Recipe Bookmarks
	// -----------------
	mux.Handle(
		"GET /recipes/bookmarks",
		optionalChain.Wrap(ListBookmarkedRecipeIds(config.Logger, config.DB)),
	)
	mux.Handle(
		"PUT /recipes/{id}/bookmark",
		protectedChain.Wrap(CreateRecipeBookmark(config.Logger, config.DB)),
	)
	mux.Handle(
		"DELETE /recipes/{id}/bookmark",
		protectedChain.Wrap(DeleteRecipeBookmark(config.Logger, config.DB)),
	)

	// -----------------
	// Recipe Reviews
	// -----------------
	mux.Handle(
		"GET /recipes/{id}/reviews",
		optionalChain.Wrap(ListRecipeReviews(config.Logger, config.DB)),
	)
	mux.Handle(
		"POST /recipes/{id}/reviews",
		protectedChain.Wrap(CreateRecipeReview(config.Logger, config.DB)),
	)
	mux.Handle(
		"DELETE /recipe-reviews/{id}",
		protectedChain.Wrap(DeleteRecipeReview(config.Logger, config.DB)),
	)

	// -----------------
	// Pokemon
	// -----------------
	mux.Handle("GET /pokemon", protectedChain.Wrap(SearchPokemon(config.Logger, config.DB)))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://julian-one.com", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Cache-Control"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	return c.Handler(mux)
}
