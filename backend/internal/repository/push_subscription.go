package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// PushSubscriptionRepository は push_subscriptions テーブルを読み書きする。
type PushSubscriptionRepository struct {
	db *gorm.DB
}

func NewPushSubscriptionRepository(db *gorm.DB) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{db: db}
}

// Upsert は購読を登録する。同じ整理券に同じ endpoint がすでにあれば、鍵を新しいものに書き換える。
// 1文で行うので、同じ購読が同時に2回送られてきても1件にしかならない。
func (r *PushSubscriptionRepository) Upsert(ctx context.Context, s *domain.PushSubscription) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ticket_id"}, {Name: "endpoint"}},
		DoUpdates: clause.AssignmentColumns([]string{"p256dh", "auth", "updated_at"}),
	}).Omit(clause.Associations).Create(s).Error
}
