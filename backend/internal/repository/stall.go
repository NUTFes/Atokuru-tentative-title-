// Package repository は GORM で DB を読み書きする。DB を触るのはこのパッケージだけにする。
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// StallRepository は stalls テーブルを読み書きする。
type StallRepository struct {
	db *gorm.DB
}

func NewStallRepository(db *gorm.DB) *StallRepository {
	return &StallRepository{db: db}
}

// FindByID は ID で屋台を1件取る。ないときは domain.ErrNotFound を返す。
func (r *StallRepository) FindByID(ctx context.Context, id uint) (*domain.Stall, error) {
	var s domain.Stall
	err := r.db.WithContext(ctx).First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateBusinessStatus は屋台の営業状況を書き換え、書き換えたあとの屋台を返す。
func (r *StallRepository) UpdateBusinessStatus(ctx context.Context, id uint, status domain.BusinessStatus) (*domain.Stall, error) {
	res := r.db.WithContext(ctx).Model(&domain.Stall{ID: id}).Update("business_status", string(status))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}
	return r.FindByID(ctx, id)
}

// StallWithWaitingCount は、屋台と、その屋台で調理中（waiting）の整理券の数。
type StallWithWaitingCount struct {
	domain.Stall `gorm:"embedded"`
	WaitingCount int
}

// ListWithWaitingCount は全ての屋台を、調理中の整理券の数と一緒に店番号の順で取る。
// 屋台ごとにクエリを投げず、LEFT JOIN と GROUP BY の1クエリでまとめて取る。
// 整理券は (stall_id, status, order_number) のインデックスで、waiting の行だけを数える。
func (r *StallRepository) ListWithWaitingCount(ctx context.Context) ([]StallWithWaitingCount, error) {
	var rows []StallWithWaitingCount
	err := r.db.WithContext(ctx).Raw(listWithWaitingCountSQL, string(domain.TicketStatusWaiting)).Scan(&rows).Error
	return rows, err
}

const listWithWaitingCountSQL = `
	SELECT stalls.*, COUNT(tickets.id) AS waiting_count
	FROM stalls
	LEFT JOIN tickets INDEXED BY idx_tickets_stall_status_order
		ON tickets.stall_id = stalls.id AND tickets.status = ?
	GROUP BY stalls.id
	ORDER BY stalls.stall_number`
