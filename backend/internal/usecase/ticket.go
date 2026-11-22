package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/repository"
)

// maxCommentLength は備考の最大の文字数（仕様書の maxLength と合わせる）。
const maxCommentLength = 200

// TicketUsecase は整理券についての処理をまとめる。
type TicketUsecase struct {
	stalls  *repository.StallRepository
	tickets *repository.TicketRepository
}

func NewTicketUsecase(stalls *repository.StallRepository, tickets *repository.TicketRepository) *TicketUsecase {
	return &TicketUsecase{stalls: stalls, tickets: tickets}
}

// Create は整理券を発行する。営業中（open）でない屋台では発行できない。
func (u *TicketUsecase) Create(ctx context.Context, stallID uint, comment *string) (*domain.Ticket, error) {
	stall, err := findStall(ctx, u.stalls, stallID)
	if err != nil {
		return nil, err
	}
	if stall.BusinessStatus != domain.BusinessStatusOpen {
		return nil, domain.StallNotOpen("営業中でない屋台では整理券を発行できません")
	}

	comment, err = normalizeComment(comment)
	if err != nil {
		return nil, err
	}

	token, err := newPublicToken()
	if err != nil {
		return nil, err
	}
	return u.tickets.Create(ctx, stallID, token, comment)
}

// normalizeComment は備考の前後の空白を取り、空なら null にする。長すぎるときはエラーにする。
func normalizeComment(comment *string) (*string, error) {
	if comment == nil {
		return nil, nil
	}
	c := strings.TrimSpace(*comment)
	if c == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(c) > maxCommentLength {
		return nil, domain.InvalidInput(fmt.Sprintf("備考は %d 文字以内にしてください", maxCommentLength))
	}
	return &c, nil
}

// newPublicToken は推測されにくいランダムな文字列（128 ビット、URL に使える22文字）を作る。
func newPublicToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("public_token を作れません: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
