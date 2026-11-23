package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/handler"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/infra/db"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/infra/notify"
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

	// 通知（Web Push）は、VAPID の鍵が3つそろっているときだけ使う。
	vapid := loadVAPID()
	var sender usecase.PushSender
	if vapid.enabled() {
		sender = notify.NewWebPush(vapid.publicKey, vapid.privateKey, vapid.subject)
	} else {
		log.Print("VAPID_PUBLIC_KEY・VAPID_PRIVATE_KEY・VAPID_SUBJECT がそろっていないので、通知（Web Push）は使いません")
	}

	stallRepo := repository.NewStallRepository(gdb)
	ticketRepo := repository.NewTicketRepository(gdb)
	pushRepo := repository.NewPushSubscriptionRepository(gdb)
	stallUsecase := usecase.NewStallUsecase(stallRepo)
	ticketUsecase := usecase.NewTicketUsecase(stallRepo, ticketRepo, pushRepo, sender)
	pushUsecase := usecase.NewPushUsecase(ticketRepo, pushRepo, vapid.enabledPublicKey())

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = handler.HTTPErrorHandler
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// API はすべて /api の下に置く（openapi.yaml の servers と合わせる）。
	h := handler.New(stallUsecase, ticketUsecase, pushUsecase)
	gen.RegisterHandlersWithBaseURL(e, gen.NewStrictHandler(h, nil), "/api")

	// /api 以外の URL では FE（Next.js の静的エクスポート）のファイルを返す。
	if dir := os.Getenv("STATIC_DIR"); dir != "" {
		handler.ServeFrontend(e, dir)
	}
	// ローカルで next dev から API を呼ぶときだけ CORS を許可する（本番は同じオリジンなので不要）。
	if origins := os.Getenv("CORS_ALLOW_ORIGINS"); origins != "" {
		handler.AllowCORS(e, strings.Split(origins, ","))
	}

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

// vapidConfig は、通知（Web Push）に使う VAPID の鍵と連絡先。
type vapidConfig struct {
	publicKey, privateKey, subject string
}

func loadVAPID() vapidConfig {
	return vapidConfig{
		publicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		privateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		subject:    os.Getenv("VAPID_SUBJECT"),
	}
}

// enabled は3つそろっているかを返す。1つでも欠けていれば通知を使わずに動かす（ローカル開発で鍵がなくても困らないように）。
func (c vapidConfig) enabled() bool {
	return c.publicKey != "" && c.privateKey != "" && c.subject != ""
}

// enabledPublicKey は、通知を使うときだけ公開鍵を返す。使わないときは空（公開鍵の API が 503 を返す）。
func (c vapidConfig) enabledPublicKey() string {
	if !c.enabled() {
		return ""
	}
	return c.publicKey
}
