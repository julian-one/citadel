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
	Admin  TestUser
	User   TestUser
	Recipe string
}

type TestUser struct {
	ID      string
	Session string
	Email   string
}

func Seed(db *sqlx.DB) *TestData {
	ctx := context.Background()

	adminID, err := user.Create(ctx, db, user.CreateRequest{
		Username: "adminuser",
		Email:    "admin@test.com",
		Password: "password123",
	})
	if err != nil {
		panic(err)
	}
	db.MustExec(`UPDATE users SET role = 'admin' WHERE user_id = ?`, adminID)

	userID, err := user.Create(ctx, db, user.CreateRequest{
		Username: "regularuser",
		Email:    "user@test.com",
		Password: "password123",
	})
	if err != nil {
		panic(err)
	}
	adminSession, err := session.Create(ctx, db, adminID)
	if err != nil {
		panic(err)
	}

	userSession, err := session.Create(ctx, db, userID)
	if err != nil {
		panic(err)
	}

	desc := "A test recipe description"
	cookTime := 5 * time.Minute
	serves := uint32(1)
	cuisine := recipe.American
	category := recipe.Main
	recipeID, err := recipe.Create(ctx, db, recipe.CreateRequest{
		User:        userID,
		Title:       "Test Recipe",
		Description: &desc,
		Components: []recipe.ComponentRequest{
			{
				Ingredients: []recipe.Ingredient{
					{Amount: 1.0, Unit: recipe.Cup, Item: "Water"},
				},
				Instructions: []string{"Boil water"},
			},
		},
		CookTime: &cookTime,
		Serves:   &serves,
		Cuisine:  &cuisine,
		Category: &category,
	})
	if err != nil {
		panic(err)
	}

	// Seed a pokemon for search tests
	db.MustExec(
		`INSERT INTO pokemon (pokemon_id, name, height, weight) VALUES (?, ?, ?, ?)`,
		1, "bulbasaur", 7, 69,
	)

	return &TestData{
		Admin: TestUser{
			ID:      adminID,
			Session: adminSession.SessionID,
			Email:   "admin@test.com",
		},
		User: TestUser{
			ID:      userID,
			Session: userSession.SessionID,
			Email:   "user@test.com",
		},
		Recipe: recipeID,
	}
}
