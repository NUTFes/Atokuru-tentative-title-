// Package domain は DB のテーブルと対応する entity と、決まっている値を置く。
package domain

import "time"

// BusinessStatus は屋台の営業状況。
type BusinessStatus string

const (
	BusinessStatusOpen      BusinessStatus = "open"      // 営業中
	BusinessStatusPreparing BusinessStatus = "preparing" // 準備中
	BusinessStatusSoldOut   BusinessStatus = "sold_out"  // 売り切れ
	BusinessStatusClosed    BusinessStatus = "closed"    // 閉店
)

// Valid は決まっている値のどれかかを返す。
func (s BusinessStatus) Valid() bool {
	switch s {
	case BusinessStatusOpen, BusinessStatusPreparing, BusinessStatusSoldOut, BusinessStatusClosed:
		return true
	}
	return false
}

// Stall は屋台（stalls テーブル）。屋台は API では作らず、テストデータと同じ方法で DB に直接入れる。
type Stall struct {
	ID uint `gorm:"primaryKey"`
	// Name は店名。
	Name string `gorm:"not null"`
	// StallNumber は店番号（A-01 など）。
	StallNumber string `gorm:"not null;uniqueIndex"`
	// PasswordHash は模擬店画面のログイン用。今はログインを使わないので null でよい。API では返さない。
	PasswordHash *string
	// HPLink は HP のリンク。ない店は null。
	HPLink *string `gorm:"column:hp_link"`
	// BusinessStatus は営業状況。
	BusinessStatus BusinessStatus `gorm:"not null;default:preparing"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
