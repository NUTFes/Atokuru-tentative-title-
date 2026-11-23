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

	for _, table := range []string{"stalls", "tickets", "comment_templates", "push_subscriptions"} {
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

// 購読のテーブルに (ticket_id, endpoint) のユニークがあり、tickets に near_notified_at の列があるかを確かめる。
func TestMigrate_PushSubscriptions(t *testing.T) {
	gdb, err := Open("file:"+filepath.Join(t.TempDir(), "test.db"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatal(err)
	}
	if !gdb.Migrator().HasIndex("push_subscriptions", "idx_push_subscriptions_ticket_endpoint") {
		t.Error("インデックス idx_push_subscriptions_ticket_endpoint がありません")
	}
	if !gdb.Migrator().HasColumn("tickets", "near_notified_at") {
		t.Error("tickets に near_notified_at の列がありません")
	}
}

// 定型文に (stall_id, comment) のユニークがあり、訳の列があるかを確かめる。
func TestMigrate_CommentTranslations(t *testing.T) {
	gdb, err := Open("file:"+filepath.Join(t.TempDir(), "test.db"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatal(err)
	}
	if !gdb.Migrator().HasIndex("comment_templates", "idx_comment_templates_stall_comment") {
		t.Error("インデックス idx_comment_templates_stall_comment がありません")
	}
	for table, col := range map[string]string{"comment_templates": "translations", "tickets": "comment_translations"} {
		if !gdb.Migrator().HasColumn(table, col) {
			t.Errorf("%s に %s の列がありません", table, col)
		}
	}
}
