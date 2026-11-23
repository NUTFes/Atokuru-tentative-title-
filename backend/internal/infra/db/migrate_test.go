package db

import (
	"path/filepath"
	"testing"
)

// テーブルとインデックスが作られるか、2回呼んでもエラーにならないかを確かめる。
func TestMigrate(t *testing.T) {
	gdb, err := Open("file:"+filepath.Join(t.TempDir(), "test.db"), "")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := Migrate(gdb); err != nil {
			t.Fatalf("%d 回目: %v", i+1, err)
		}
	}

	for _, table := range []string{"stalls", "tickets", "comment_templates"} {
		if !gdb.Migrator().HasTable(table) {
			t.Errorf("テーブル %s がありません", table)
		}
	}

	var indexes []string
	if err := gdb.Raw("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = 'tickets'").Scan(&indexes).Error; err != nil {
		t.Fatal(err)
	}
	has := map[string]bool{}
	for _, name := range indexes {
		has[name] = true
	}
	for _, want := range []string{"idx_tickets_stall_order", "idx_tickets_stall_status_order", "idx_tickets_public_token"} {
		if !has[want] {
			t.Errorf("インデックス %s がありません（あるもの: %v）", want, indexes)
		}
	}
}
