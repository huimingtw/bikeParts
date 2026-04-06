package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huimingtw/bikeparts/db"
	"github.com/huimingtw/bikeparts/handler"
	"github.com/huimingtw/bikeparts/service"

	"github.com/gin-gonic/gin"
)

func main() {
	database, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	router := gin.Default()

	mailer := &service.EmailServiceImpl{}
	h := handler.NewHandler(database, mailer)
	router.GET("/api/mail_test", h.MailTest)

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
