package test

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"flag"

	"citadel/route"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	server *httptest.Server
	td     *TestData
)

func TestMain(m *testing.M) {
	flag.Parse()

	db := sqlx.MustConnect("sqlite3", ":memory:?_foreign_keys=on")

	schemaSQL, err := os.ReadFile(filepath.Join("..", "schema", "model.sql"))
	if err != nil {
		panic(err)
	}
	db.MustExec(string(schemaSQL))

	// Only log if the test is run with the -v flag
	var logOutput io.Writer = io.Discard
	if testing.Verbose() {
		logOutput = os.Stdout
	}
	logger := slog.New(slog.NewJSONHandler(logOutput, nil))

	// Initialize the server with the test database and logger
	handler := route.Initialize(route.Config{
		Db:     db,
		Logger: logger,
	})
	server = httptest.NewServer(handler)

	// Seed the database with test data
	td = Seed(db)

	// Run the tests
	code := m.Run()

	// NOTE: defer doesn't work here because os.Exit will terminate the program immediately
	server.Close()
	db.Close()

	os.Exit(code)
}
