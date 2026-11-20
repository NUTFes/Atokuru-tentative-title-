// Package handler は gen.StrictServerInterface を実装する。
// リクエストを受け取って usecase を呼び、結果を返すだけにする。
package handler

import "github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"

// Handler は全 API の実装をまとめる。usecase は API を追加するときにフィールドとして足す。
type Handler struct{}

// コンパイル時に、仕様書の全 API を実装しているかを確かめる。
var _ gen.StrictServerInterface = (*Handler)(nil)

func New() *Handler {
	return &Handler{}
}
