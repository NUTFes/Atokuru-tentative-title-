// Package handler は gen.StrictServerInterface を実装する。
// リクエストを受け取って usecase を呼び、結果を返すだけにする。
package handler

import (
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/usecase"
)

// Handler は全 API の実装をまとめる。
type Handler struct {
	stalls  *usecase.StallUsecase
	tickets *usecase.TicketUsecase
}

// コンパイル時に、仕様書の全 API を実装しているかを確かめる。
var _ gen.StrictServerInterface = (*Handler)(nil)

func New(stalls *usecase.StallUsecase, tickets *usecase.TicketUsecase) *Handler {
	return &Handler{stalls: stalls, tickets: tickets}
}
