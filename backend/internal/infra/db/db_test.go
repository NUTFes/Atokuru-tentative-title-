package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"
)

// connCheck は接続確認のためだけに使うテーブル。テストの最後に消す。
type connCheck struct {
	ID        uint `gorm:"primaryKey"`
	Message   string
	CreatedAt time.Time
}

func (connCheck) TableName() string { return "conn_check_tmp" }

// checkReadWrite は、テーブルの作成・書き込み・読み込み・更新・削除ができるかを確かめる。
func checkReadWrite(t *testing.T, gdb *gorm.DB) {
	t.Helper()

	if err := gdb.AutoMigrate(&connCheck{}); err != nil {
		t.Fatalf("テーブルを作れません: %v", err)
	}
	t.Cleanup(func() {
		if err := gdb.Migrator().DropTable(&connCheck{}); err != nil {
			t.Errorf("テーブルを消せません: %v", err)
		}
	})

	row := connCheck{Message: "hello"}
	if err := gdb.Create(&row).Error; err != nil {
		t.Fatalf("書き込めません: %v", err)
	}

	var got connCheck
	if err := gdb.First(&got, row.ID).Error; err != nil {
		t.Fatalf("読み込めません: %v", err)
	}
	if got.Message != "hello" {
		t.Fatalf("読み込んだ値が違います: got %q, want %q", got.Message, "hello")
	}

	if err := gdb.Model(&got).Update("message", "world").Error; err != nil {
		t.Fatalf("更新できません: %v", err)
	}
	if err := gdb.First(&got, row.ID).Error; err != nil || got.Message != "world" {
		t.Fatalf("更新した値が読めません: got %q, err %v", got.Message, err)
	}

	if err := gdb.Delete(&got).Error; err != nil {
		t.Fatalf("削除できません: %v", err)
	}
}

// 手元の SQLite ファイルで試す。Turso の設定がなくても動く。
func TestOpen_LocalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	gdb, err := Open("file:"+path, "")
	if err != nil {
		t.Fatal(err)
	}
	checkReadWrite(t, gdb)
}

// 本物の Turso で試す。環境変数 TURSO_DATABASE_URL がなければスキップする。
func TestOpen_Turso(t *testing.T) {
	url := os.Getenv("TURSO_DATABASE_URL")
	if url == "" {
		t.Skip("TURSO_DATABASE_URL が設定されていないのでスキップします")
	}
	gdb, err := Open(url, os.Getenv("TURSO_AUTH_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}
	checkReadWrite(t, gdb)
}
