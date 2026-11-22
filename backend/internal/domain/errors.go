package domain

import "errors"

// エラーの種類。usecase はこれを包んだ *Error を返し、handler が HTTP のステータスと code に変える。
var (
	ErrNotFound     = errors.New("not_found")      // 指定したものがない
	ErrInvalidInput = errors.New("bad_request")    // 入力の値が正しくない
	ErrStallNotOpen = errors.New("stall_not_open") // 営業中でない屋台で発券しようとした
)

// Error は、種類（Kind）と、利用者に見せるメッセージを持つエラー。
type Error struct {
	Kind    error
	Message string
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Kind }

func NotFound(message string) error     { return &Error{Kind: ErrNotFound, Message: message} }
func InvalidInput(message string) error { return &Error{Kind: ErrInvalidInput, Message: message} }
func StallNotOpen(message string) error { return &Error{Kind: ErrStallNotOpen, Message: message} }
