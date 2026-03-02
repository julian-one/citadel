package test

import (
	"context"
	"time"

	"citadel/internal/recipe"
	"citadel/internal/session"
	"citadel/internal/user"

	"github.com/jmoiron/sqlx"
)

type TestData struct {
	Admin    TestUser
	User     TestUser
	RecipeId string
}

type TestUser struct {
	Id      string
	Session string
	Email   string
}

func Seed(db *sqlx.DB) *TestData {
	ctx := context.Background()

	adminId, err := user.Create(ctx, db, user.CreateRequest{
		Username: "adminuser",
		Email:    "admin@test.com",
		Password: "password123",
	})
	if err != nil {
		panic(err)
	}
	db.MustExec(`UPDATE users SET role = 'admin' WHERE user_id = ?`, adminId)

	userId, err := user.Create(ctx, db, user.CreateRequest{
		Username: "regularuser",
		Email:    "user@test.com",
		Password: "password123",
	})
	if err != nil {
		panic(err)
	}

	adminSession, err := session.New(ctx, db, adminId)
	if err != nil {
		panic(err)
	}

	userSession, err := session.New(ctx, db, userId)
	if err != nil {
		panic(err)
	}

	desc := "A test recipe description"
	cookTime := 5 * time.Minute
	serves := uint32(1)
	cuisine := recipe.American
	category := recipe.Main
	recipeId, err := recipe.Create(ctx, db, recipe.CreateRequest{
		User:        userId,
		Title:       "Test Recipe",
		Description: &desc,
		Ingredients: []recipe.Ingredient{
			{Amount: 1.0, Unit: recipe.Cup, Item: "Water"},
		},
		Instructions: []string{"Boil water"},
		CookTime:     &cookTime,
		Serves:       &serves,
		Cuisine:      &cuisine,
		Category:     &category,
	})
	if err != nil {
		panic(err)
	}

	return &TestData{
		Admin: TestUser{
			Id:      adminId,
			Session: adminSession.SessionId,
			Email:   "admin@test.com",
		},
		User: TestUser{
			Id:      userId,
			Session: userSession.SessionId,
			Email:   "user@test.com",
		},
		RecipeId: recipeId,
	}
}
