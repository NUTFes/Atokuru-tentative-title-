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

// MaxReceivedOrderNumber は、その屋台で受け取り済み（received）の整理券の最大の order_number を返す。
// 1枚もなければ 0 を返す。(stall_id, status, order_number) のインデックスで、数行だけ読んで済む。
func (r *TicketRepository) MaxReceivedOrderNumber(ctx context.Context, stallID uint) (int, error) {
	var maxNum int
	err := r.db.WithContext(ctx).Raw(
		"SELECT COALESCE(MAX(order_number), 0) FROM tickets WHERE stall_id = ? AND status = ?",
		stallID, string(domain.TicketStatusReceived),
	).Scan(&maxNum).Error
	return maxNum, err
}

// ListByStall は、その屋台の整理券のうち status が statuses のどれかのものを、order_number の順に最大 limit 件取る。
func (r *TicketRepository) ListByStall(ctx context.Context, stallID uint, statuses []domain.TicketStatus, limit int, descending bool) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	err := r.listByStallQuery(ctx, stallID, statuses, limit, descending).Find(&tickets).Error
	return tickets, err
}

// listByStallQuery は ListByStall のクエリを組み立てる（テストで EXPLAIN にかけるために分けている）。
//
// status を複数指定すると、SQLite は並べ替えを省くために (stall_id, order_number) のインデックスを選び、
// 受け取り済みも含めた屋台の整理券を全部たどってしまう（10万件・ANALYZE 後でも同じ）。
// Turso は読んだ行数で無料枠を消費するので、INDEXED BY で (stall_id, status, order_number) を使わせ、
// 指定した status の行だけを読むようにする。
func (r *TicketRepository) listByStallQuery(ctx context.Context, stallID uint, statuses []domain.TicketStatus, limit int, descending bool) *gorm.DB {
	order := "order_number ASC"
	if descending {
		order = "order_number DESC"
	}
	return r.db.WithContext(ctx).
		Table("tickets INDEXED BY idx_tickets_stall_status_order").
		Where("stall_id = ? AND status IN ?", stallID, statuses).
		Order(order).
		Limit(limit)
}
