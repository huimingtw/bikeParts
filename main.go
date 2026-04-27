package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/huimingtw/bikeparts/config"
	"github.com/huimingtw/bikeparts/db"
	"github.com/huimingtw/bikeparts/handler"
	"github.com/huimingtw/bikeparts/middleware"
	"github.com/huimingtw/bikeparts/router"
	"github.com/huimingtw/bikeparts/service"

	"github.com/joho/godotenv"
	"gopkg.in/lumberjack.v2"
)

//go:embed db/schema.sql frontend/*
var assetsFS embed.FS

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Another instance is already running — open the browser and exit without starting a second server.
	port := cfg.Get().Port
	if port == "" {
		port = "8080"
	}
	if isAlreadyRunning(port) {
		openBrowser("http://localhost:" + port)
		return
	}

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

	logWriter := openLogFile()
	logger := slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, logWriter), &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	mailer := service.NewEmailService(cfg)
	notifier := service.NewNotificationService(mailer, logger, database, func() bool {
		return cfg.Get().EmailNotificationsEnabled
	})

	// Start a background goroutine to retry sending unsent notifications every day at noon.
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
			if !next.After(now) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next))
			notifier.RetryUnsent(context.Background())
		}
	}()

	h := handler.NewHandler(database, notifier, logger, cfg, mailer)
	ic := middleware.NewIdempotencyCache()

	frontendFS, err := fs.Sub(assetsFS, "frontend")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}
	r := router.NewRouter(h, ic, logger, frontendFS)

	settings := cfg.Get()
	PORT := settings.Port
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
	runTray(server, PORT)
}

func openLogFile() io.Writer {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return io.Discard
	}
	logPath := filepath.Join(baseDir, "bikeparts", "app.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return io.Discard
	}
	return &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // MB
		MaxBackups: 3,
		MaxAge:     30, // days
		Compress:   true,
	}
}

func isAlreadyRunning(port string) bool {
	conn, err := net.DialTimeout("tcp", "localhost:"+port, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func defaultDBPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %v", err)
	}
	return fmt.Sprintf("%s/bikeparts/data.db", baseDir), nil
}
