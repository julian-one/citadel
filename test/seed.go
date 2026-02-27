package test

import (
	"context"

	"citadel/internal/session"
	"citadel/internal/user"

	"github.com/jmoiron/sqlx"
)

type TestData struct {
	Admin TestUser
	User  TestUser
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
	}
}
