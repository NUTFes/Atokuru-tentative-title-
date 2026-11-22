package repository

import (
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/infra/db"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := db.Open("file:"+filepath.Join(t.TempDir(), "test.db"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(gdb); err != nil {
		t.Fatal(err)
	}
	return gdb
}

// assertUsesIndex は、クエリが全件を読まず（SCAN tickets にならず）、インデックスを使うことを確かめる。
func assertUsesIndex(t *testing.T, gdb *gorm.DB, index, query string, args ...any) {
	t.Helper()
	rows, err := gdb.Raw("EXPLAIN QUERY PLAN "+query, args...).Rows()
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var plan []string
	for rows.Next() {
		var id, parent, notused int
		var detail string
		if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
			t.Fatal(err)
		}
		plan = append(plan, detail)
	}
	joined := strings.Join(plan, " / ")
	if !strings.Contains(joined, index) {
		t.Errorf("インデックス %s を使っていません: %s", index, joined)
	}
	for _, p := range plan {
		if strings.HasPrefix(p, "SCAN tickets") && !strings.Contains(p, "USING") {
			t.Errorf("tickets を全件読んでいます: %s", joined)
		}
	}
}
