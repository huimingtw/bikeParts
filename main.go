package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/huimingtw/bikeparts/db"
	"github.com/huimingtw/bikeparts/handler"
	"github.com/huimingtw/bikeparts/middleware"
	"github.com/huimingtw/bikeparts/router"
	"github.com/huimingtw/bikeparts/service"

	"github.com/joho/godotenv"
)

//go:embed db/schema.sql frontend/*
var assetsFS embed.FS

func main() {
	_ = godotenv.Load()

	schemaBytes, err := assetsFS.ReadFile("db/schema.sql")
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath, err = defaultDBPath()
		if err != nil {
			log.Fatalf("Failed to determine default DB path: %v", err)
		}
	}

	database, err := db.Init(db.Config{
		DBPath: dbPath,
		Schema: schemaBytes,
	})
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	mailer := service.NewEmailService()
	notifier := service.NewNotificationService(mailer, logger)
	h := handler.NewHandler(database, notifier, logger)
	ic := middleware.NewIdempotencyCache()

	frontendFS, err := fs.Sub(assetsFS, "frontend")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	r := router.NewRouter(h, ic, logger, http.FS(frontendFS))

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server is running on port %s", PORT)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

}

func defaultDBPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %v", err)
	}
	return filepath.Join(baseDir, "bikeparts", "data.db"), nil
}
