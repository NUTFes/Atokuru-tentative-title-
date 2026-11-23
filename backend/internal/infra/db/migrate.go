package db

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// Migrate は domain の entity に合わせてテーブルとインデックスを作る。
// サーバーの起動時に毎回呼ぶ。足りないテーブル・カラム・インデックスだけを足すので、何度呼んでもよい。
// （GORM の AutoMigrate はカラムの削除や型の変更まではしない。必要になったら手で直す）
func Migrate(gdb *gorm.DB) error {
	if err := gdb.AutoMigrate(
		&domain.Stall{},
		&domain.Ticket{},
		&domain.CommentTemplate{},
	); err != nil {
		return fmt.Errorf("テーブルを作れません: %w", err)
	}
	return nil
}
