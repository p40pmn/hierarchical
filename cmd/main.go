package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phuangpheth/hierarchical"
	"github.com/phuangpheth/hierarchical/database"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func main() {
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	timeZone := getEnv(os.Getenv("TZ"), "Asia/Vientiane")
	dbConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s", dbHost, dbPort, dbUser, dbPass, dbName, timeZone)

	db, err := database.Open(getEnv("DB_DRIVER", "postgres"), dbConn)
	failOnError(err, "failed to open database")
	defer func() {
		if err := db.Close(); err != nil {
			failOnError(err, "failed to close database")
		}
	}()
	if err := db.Ping(context.Background()); err != nil {
		failOnError(err, "failed to ping database")
	}

	svc := hierarchical.NewService(db)
	handler := newHandler(svc)

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/v1/syllabuses/:id", handler.GetByID)

	go func() {
		if err := e.Start(fmt.Sprintf(":%s", getEnv("PORT", "8080"))); err != nil {
			e.Logger.Fatal("Shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutdown in progress...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal("Failed to shutdown the server", err)
	}
}

func newHandler(service *hierarchical.Service) *handler {
	return &handler{service: service}
}

type handler struct {
	service *hierarchical.Service
}

func (h *handler) GetByID(c echo.Context) error {
	syllabus, err := h.service.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, syllabus)
}
