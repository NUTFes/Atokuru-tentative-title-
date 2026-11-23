// seed は、屋台の CSV と整理券のテストデータを DB に入れるコマンド。
//
// backend/ で、.env の環境変数を読み込んでから実行する。
//
//	set -a; . ./.env; set +a
//	go run ./cmd/seed                          # testdata/stalls.csv を入れる
//	go run ./cmd/seed -csv path/to/stalls.csv  # 別の CSV を入れる
//	go run ./cmd/seed -tickets 100000          # 屋台を入れたあと、整理券を10万件作って入れる
//	go run ./cmd/seed -tickets 100000 -reset   # 今ある整理券を全部消してから入れ直す
//
// 整理券はプログラムで作る（ファイルにはしない）。-seed が同じなら、誰が入れても同じデータになる。
// 入れる前に接続先を表示して確認する。-yes を付けると確認しない。
// テストデータには <script> を含む店名などがあり、public_token も seed から分かるので、本番の DB には入れないこと。
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/infra/db"
)

func main() {
	csvPath := flag.String("csv", "testdata/stalls.csv", "入れる屋台の CSV")
	yes := flag.Bool("yes", false, "接続先の確認をしない")
	ticketCount := flag.Int("tickets", 0, "作って入れる整理券の件数（0 なら整理券は入れない）")
	seed := flag.Uint64("seed", 42, "整理券を作る乱数の seed")
	reset := flag.Bool("reset", false, "整理券がすでにあるとき、全部消してから入れ直す")
	flag.Parse()
	if *ticketCount < 0 {
		log.Fatal("-tickets は 0 以上にしてください")
	}
	if *reset && *ticketCount == 0 {
		log.Fatal("-reset は -tickets と一緒に使ってください")
	}

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
	if *ticketCount > 0 {
		fmt.Printf("整理券を %d 件作って入れます（seed %d", *ticketCount, *seed)
		if *reset {
			fmt.Print("、今ある整理券は全部消します")
		}
		fmt.Println("）")
	}
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

	if *ticketCount > 0 {
		if err := seedTickets(gdb, *ticketCount, *seed, *reset); err != nil {
			log.Fatal(err)
		}
	}
}

// seedTickets は、DB にある屋台に整理券を割り振って入れ、集計だけを表示する。
func seedTickets(gdb *gorm.DB, count int, seed uint64, reset bool) error {
	dbStalls, err := db.ListStalls(gdb)
	if err != nil {
		return err
	}
	tickets, err := db.GenerateTickets(dbStalls, count, seed)
	if err != nil {
		return err
	}

	start := time.Now()
	const reportEvery = 10000
	err = db.SeedTickets(gdb, tickets, reset, func(done int) {
		if done%reportEvery == 0 || done == len(tickets) {
			fmt.Printf("  %d / %d 件（%s）\n", done, len(tickets), time.Since(start).Round(time.Second))
		}
	})
	if err != nil {
		return err
	}

	sum, err := db.SummarizeTickets(gdb)
	if err != nil {
		return err
	}
	fmt.Printf("整理券を入れました: waiting %d / ready %d / received %d\n",
		sum.ByStatus[domain.TicketStatusWaiting], sum.ByStatus[domain.TicketStatusReady], sum.ByStatus[domain.TicketStatusReceived])

	// 混雑度の閾値（usecase）と同じ区切りで、営業中の屋台の待ち人数を数える。
	var green, yellow, red, notOpen int
	for _, s := range sum.Stalls {
		switch {
		case s.BusinessStatus != domain.BusinessStatusOpen:
			notOpen++
		case s.Waiting >= 10:
			red++
		case s.Waiting >= 5:
			yellow++
		default:
			green++
		}
	}
	fmt.Printf("屋台の待ち人数: 営業中で0〜4人 %d 店 / 5〜9人 %d 店 / 10人以上 %d 店、営業していない店 %d 店\n", green, yellow, red, notOpen)
	return nil
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
