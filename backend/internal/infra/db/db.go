// Package db は Turso（libSQL）への接続を作る。
//
// GORM は Turso に標準対応していないので、libSQL のドライバ（libsql-client-go）で
// database/sql の接続を作り、それを GORM の SQLite ドライバに渡して使う。
package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Open は url の DB に接続する。
//
//   - Turso に繋ぐとき：url は libsql://<DB名>-<組織名>.turso.io、authToken は turso CLI で発行したトークン
//   - 手元のファイルで試すとき：url は file:local.db のようにし、authToken は空にする
func Open(url, authToken string) (*gorm.DB, error) {
	if url == "" {
		return nil, errors.New("DB の URL が空です（環境変数 TURSO_DATABASE_URL を設定してください）")
	}

	var opts []libsql.Option
	if authToken != "" {
		opts = append(opts, libsql.WithAuthToken(authToken))
	}
	connector, err := libsql.NewConnector(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("libSQL の接続設定を作れません: %w", err)
	}
	sqlDB := sql.OpenDB(connector)

	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("GORM で DB を開けません: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("DB に接続できません: %w", err)
	}
	return gormDB, nil
}
