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
