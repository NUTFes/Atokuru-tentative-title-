package domain

import "time"

// CommentTemplate は備考の定型文（comment_templates テーブル）。
// テーブルだけ先に作っておく。読み書きする API は後回し。
type CommentTemplate struct {
	ID        uint   `gorm:"primaryKey"`
	StallID   uint   `gorm:"not null;index"`
	Stall     Stall  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comment   string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
