package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huimingtw/bikeparts/db"
	"github.com/huimingtw/bikeparts/handler"
	"github.com/huimingtw/bikeparts/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	database, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	router := gin.Default()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	mailer := service.NewEmailService()
	notifier := service.NewNotificationService(mailer, logger)
	h := handler.NewHandler(database, notifier, logger)

	api := router.Group("/api")
	api.GET("/mail_test", h.MailTest)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: router,
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
