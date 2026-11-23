// seed は、屋台の CSV を DB に入れるコマンド。
//
// backend/ で、.env の環境変数を読み込んでから実行する。
//
//	set -a; . ./.env; set +a
//	go run ./cmd/seed                          # testdata/stalls.csv を入れる
//	go run ./cmd/seed -csv path/to/stalls.csv  # 別の CSV を入れる
//
// 入れる前に接続先を表示して確認する。-yes を付けると確認しない。
// テストデータには <script> を含む店名などがあるので、本番の DB には入れないこと。
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/infra/db"
)

func main() {
	csvPath := flag.String("csv", "testdata/stalls.csv", "入れる屋台の CSV")
	yes := flag.Bool("yes", false, "接続先の確認をしない")
	flag.Parse()

	f, err := os.Open(*csvPath)
	if err != nil {
		log.Fatal(err)
	}
	stalls, err := db.ReadStallsCSV(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	dbURL := os.Getenv("TURSO_DATABASE_URL")
	fmt.Printf("接続先: %s\n", describeTarget(dbURL))
	fmt.Printf("%s の %d 店を入れます（店番号が同じ屋台は上書きします）\n", *csvPath, len(stalls))
	if !*yes && !confirm() {
		fmt.Println("やめました")
		return
	}

	gdb, err := db.Open(dbURL, os.Getenv("TURSO_AUTH_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Migrate(gdb); err != nil {
		log.Fatal(err)
	}
	if err := db.SeedStalls(gdb, stalls); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d 店を入れました\n", len(stalls))
}

// describeTarget は、接続先を人が確かめられる形にする。トークンなどのクエリは表示しない。
func describeTarget(dbURL string) string {
	u, err := url.Parse(dbURL)
	if err != nil || dbURL == "" {
		return "（TURSO_DATABASE_URL が読めません）"
	}
	if u.Scheme == "file" {
		return "手元のファイル " + strings.TrimPrefix(u.Opaque+u.Path, "//")
	}
	return u.Scheme + "://" + u.Host
}

func confirm() bool {
	fmt.Print("続けますか？ [y/N]: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(strings.ToLower(line)) == "y"
}
