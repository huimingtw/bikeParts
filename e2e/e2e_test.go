package e2e

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/huimingtw/bikeparts/db"
	"github.com/huimingtw/bikeparts/config"
	"github.com/huimingtw/bikeparts/handler"
	"github.com/huimingtw/bikeparts/middleware"
	"github.com/huimingtw/bikeparts/router"
	"github.com/huimingtw/bikeparts/service"
)

type noopMailer struct{}

func (n *noopMailer) Send(subject, body string) error { return nil }

var testRouter *gin.Engine
var testDB *sql.DB

func TestMain(m *testing.M) {
	// create in-memory SQLite database
	schemaBytes, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		panic(err)
	}

	db, err := sqlite.Init(sqlite.Config{
		DBPath: ":memory:",
		Schema: schemaBytes,
	})
	if err != nil {
		panic(err)
	}

	// initialize router with test dependencies
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	mailer := &noopMailer{}
	notifier := service.NewNotificationService(mailer, logger)
	cfg := config.NewInMemory()
	h := handler.NewHandler(db, notifier, logger, cfg, mailer)
	ic := middleware.NewIdempotencyCache()

	// expose resources
	testRouter = router.NewRouter(h, ic, logger, os.DirFS("../frontend"))
	testDB = db

	os.Exit(m.Run())
}

func makeRequest(method, url string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func truncateTables() {
	tables := []string{
		"parts",
		"stock_movements",
		"low_stock_notifications",
	}
	for _, t := range tables {
		testDB.Exec("DELETE FROM " + t)
		testDB.Exec("DELETE FROM sqlite_sequence WHERE name = ?", t)
	}
}
