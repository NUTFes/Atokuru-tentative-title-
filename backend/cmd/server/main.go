package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/handler"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/infra/db"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/repository"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/usecase"
)

func main() {
	gdb, err := db.Open(os.Getenv("TURSO_DATABASE_URL"), os.Getenv("TURSO_AUTH_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Migrate(gdb); err != nil {
		log.Fatal(err)
	}

	stallRepo := repository.NewStallRepository(gdb)
	ticketRepo := repository.NewTicketRepository(gdb)
	stallUsecase := usecase.NewStallUsecase(stallRepo)
	ticketUsecase := usecase.NewTicketUsecase(stallRepo, ticketRepo)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = handler.HTTPErrorHandler
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// API はすべて /api の下に置く（openapi.yaml の servers と合わせる）。
	// それ以外の URL は、FE の静的ファイルを返すために空けておく。
	h := handler.New(stallUsecase, ticketUsecase)
	gen.RegisterHandlersWithBaseURL(e, gen.NewStrictHandler(h, nil), "/api")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		if err := e.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal(err)
		}
	}()

	// Ctrl+C やコンテナの停止で、処理中のリクエストを終えてから止める。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Fatal(err)
	}
}
