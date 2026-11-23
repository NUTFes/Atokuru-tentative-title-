package domain

import "time"

// PushSubscription は、整理券に紐づいた通知（Web Push）の購読（push_subscriptions テーブル）。
// ブラウザの PushSubscription のうち、通知を送るのに必要な endpoint と鍵だけを持つ。
//
// 1つの整理券に複数の端末が、1つの端末に複数の整理券が紐づくことがある。
// (ticket_id, endpoint) のユニークで、同じ端末を同じ整理券に二重に登録しないようにする。
type PushSubscription struct {
	ID       uint   `gorm:"primaryKey"`
	TicketID uint   `gorm:"not null;uniqueIndex:idx_push_subscriptions_ticket_endpoint,priority:1"`
	Ticket   Ticket `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Endpoint はプッシュサーバーの URL。通知はこの URL に送る。
	Endpoint string `gorm:"not null;uniqueIndex:idx_push_subscriptions_ticket_endpoint,priority:2"`
	// P256dh と Auth は、通知の中身を暗号化するための鍵（ブラウザが作る。base64url）。
	P256dh    string `gorm:"column:p256dh;not null"`
	Auth      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
