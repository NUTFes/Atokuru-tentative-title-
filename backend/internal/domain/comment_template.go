package domain

import "time"

// Translations は文の訳。キーは言語コード（FE の locales/ のファイル名と同じ。例：en）、値は訳した文。
type Translations map[string]string

// CommentTemplate は備考の定型文（comment_templates テーブル）。
// 登録したときに一度だけ訳して Translations に保存する（来場者が見るたびには訳さない）。
//
// (stall_id, comment) の複合ユニークで、同じ屋台に同じ文を二重に登録しないようにする。
// 発券したときに、備考と同じ定型文を探すのにも使う。
type CommentTemplate struct {
	ID      uint   `gorm:"primaryKey"`
	StallID uint   `gorm:"not null;uniqueIndex:idx_comment_templates_stall_comment,priority:1"`
	Stall   Stall  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comment string `gorm:"not null;uniqueIndex:idx_comment_templates_stall_comment,priority:2"`
	// Translations は Comment の訳（JSON で保存する）。訳せなかったときは空。
	Translations Translations `gorm:"type:text;serializer:json"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
