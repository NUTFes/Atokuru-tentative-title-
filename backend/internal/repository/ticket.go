package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// TicketRepository は tickets テーブルを読み書きする。
type TicketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// createRetries は、同時に発券されて order_number がぶつかったときにやり直す回数。
const createRetries = 5

// Create は整理券を1件作る。order_number は、その屋台の最大の番号 + 1 にする。
//
// 番号を決めるのと書き込むのを1つの INSERT 文で行うので、SQLite（Turso）が書き込みを1つずつ
// 順番に処理する限り、同時に発券されても番号は重複しない。万一ぶつかっても、
// (stall_id, order_number) の複合ユニークで弾かれるので、そのときはやり直す。
func (r *TicketRepository) Create(ctx context.Context, stallID uint, publicToken string, comment *string) (*domain.Ticket, error) {
	var lastErr error
	for i := 0; i < createRetries; i++ {
		now := time.Now()
		err := r.db.WithContext(ctx).Exec(`
			INSERT INTO tickets (public_token, stall_id, order_number, status, comment, created_at, updated_at)
			SELECT ?, ?, COALESCE(MAX(order_number), 0) + 1, ?, ?, ?, ?
			FROM tickets WHERE stall_id = ?`,
			publicToken, stallID, string(domain.TicketStatusWaiting), comment, now, now, stallID,
		).Error
		if err == nil {
			return r.FindByPublicToken(ctx, publicToken)
		}
		if !isOrderNumberConflict(err) {
			return nil, err
		}
		lastErr = err
	}
	return nil, fmt.Errorf("order_number がぶつかり続けたので発券できません: %w", lastErr)
}

func isOrderNumberConflict(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") && strings.Contains(msg, "order_number")
}

// FindByPublicToken は public_token で整理券を1件取る。ないときは domain.ErrNotFound を返す。
func (r *TicketRepository) FindByPublicToken(ctx context.Context, publicToken string) (*domain.Ticket, error) {
	var t domain.Ticket
	err := r.db.WithContext(ctx).Where("public_token = ?", publicToken).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
