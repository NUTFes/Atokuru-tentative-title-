package domain

import "time"

// TicketStatus は整理券の状態。
type TicketStatus string

const (
	TicketStatusWaiting  TicketStatus = "waiting"  // 調理中
	TicketStatusReady    TicketStatus = "ready"    // 調理完了
	TicketStatusReceived TicketStatus = "received" // 受け取り済み
)

// Valid は決まっている値のどれかかを返す。
func (s TicketStatus) Valid() bool {
	switch s {
	case TicketStatusWaiting, TicketStatusReady, TicketStatusReceived:
		return true
	}
	return false
}

// Ticket は整理券（tickets テーブル）。
//
// インデックス：
//   - (stall_id, order_number) の複合ユニーク：同じ屋台で番号が重複しないようにする
//   - (stall_id, status, order_number)：待ち一覧の取得と、前に何人の計算（received の最大番号）を速くする
//   - public_token のユニーク
type Ticket struct {
	ID uint `gorm:"primaryKey"`
	// PublicToken は推測されにくいランダムな文字列。QR コードの URL にはこれを入れる（id は連番なので使わない）。
	PublicToken string `gorm:"not null;uniqueIndex"`
	StallID     uint   `gorm:"not null;uniqueIndex:idx_tickets_stall_order,priority:1;index:idx_tickets_stall_status_order,priority:1"`
	Stall       Stall  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	// OrderNumber は屋台ごとの整理券番号（1 から）。
	OrderNumber int          `gorm:"not null;uniqueIndex:idx_tickets_stall_order,priority:2;index:idx_tickets_stall_status_order,priority:3"`
	Status      TicketStatus `gorm:"not null;default:waiting;index:idx_tickets_stall_status_order,priority:2"`
	// Comment は備考（メニューなど）。ない場合は null。
	Comment *string
	// CalledAt は調理完了（ready）にした時刻。まだなら null。
	CalledAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
