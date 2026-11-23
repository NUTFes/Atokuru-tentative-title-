package db

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// testTicketCount は、テストで作る整理券の数。10万件だと遅いので減らす（性質は件数によらない）。
const testTicketCount = 3000

// testdataStalls は testdata/stalls.csv の80店に、DB に入れたときのように ID を振ったもの。
func testdataStalls(t *testing.T) []domain.Stall {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "..", "testdata", "stalls.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	stalls, err := ReadStallsCSV(f)
	if err != nil {
		t.Fatal(err)
	}
	for i := range stalls {
		stalls[i].ID = uint(i + 1)
	}
	return stalls
}

func TestGenerateTickets(t *testing.T) {
	stalls := testdataStalls(t)
	tickets, err := GenerateTickets(stalls, testTicketCount, 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(tickets) != testTicketCount {
		t.Fatalf("整理券が %d 件です（%d 件のはず）", len(tickets), testTicketCount)
	}

	statusOf := map[uint]domain.BusinessStatus{}
	for _, s := range stalls {
		statusOf[s.ID] = s.BusinessStatus
	}

	tokens := map[string]bool{}
	byStall := map[uint][]domain.Ticket{}
	byStatus := map[domain.TicketStatus]int{}
	for _, tk := range tickets {
		if len(tk.PublicToken) != 22 || tokens[tk.PublicToken] {
			t.Fatalf("public_token が22文字でないか、重複しています: %q", tk.PublicToken)
		}
		tokens[tk.PublicToken] = true
		if !tk.Status.Valid() {
			t.Fatalf("status %q は決まっている値ではありません", tk.Status)
		}
		byStatus[tk.Status]++
		if tk.Comment != nil && utf8.RuneCountInString(*tk.Comment) > 200 {
			t.Errorf("備考が200文字を超えています: %d 文字", utf8.RuneCountInString(*tk.Comment))
		}
		switch {
		case tk.Status == domain.TicketStatusWaiting && tk.CalledAt != nil:
			t.Errorf("waiting なのに called_at があります: %+v", tk)
		case tk.Status != domain.TicketStatusWaiting && (tk.CalledAt == nil || tk.CalledAt.Before(tk.CreatedAt)):
			t.Errorf("%s なのに called_at がないか、発券より前です: %+v", tk.Status, tk)
		}
		byStall[tk.StallID] = append(byStall[tk.StallID], tk)
	}
	for _, st := range []domain.TicketStatus{domain.TicketStatusWaiting, domain.TicketStatusReady, domain.TicketStatusReceived} {
		if byStatus[st] == 0 {
			t.Errorf("status %s の整理券がありません", st)
		}
	}

	// 混雑度の色ごとの、営業中の屋台の数（0〜4人で緑、5〜9人で黄、10人以上で赤）。
	colors := map[string]int{}
	staleStalls := 0
	for _, s := range stalls {
		list := byStall[s.ID]
		if len(list) == 0 {
			t.Errorf("屋台 %s に整理券がありません", s.StallNumber)
			continue
		}

		waiting, maxReceived, stale := 0, 0, false
		for j, tk := range list {
			if tk.OrderNumber != j+1 {
				t.Fatalf("屋台 %s の番号が1からの連番ではありません: %d 枚目が %d 番", s.StallNumber, j+1, tk.OrderNumber)
			}
			if j > 0 && tk.CreatedAt.Before(list[j-1].CreatedAt) {
				t.Errorf("屋台 %s の %d 番が、前の番号より先に発券されています", s.StallNumber, tk.OrderNumber)
			}
			switch tk.Status {
			case domain.TicketStatusWaiting:
				waiting++
			case domain.TicketStatusReceived:
				maxReceived = tk.OrderNumber
			}
		}
		if maxReceived == 0 {
			t.Errorf("屋台 %s に received の整理券がありません", s.StallNumber)
		}
		// received の最大番号より後ろに waiting が並び、received より前に waiting がない。
		for _, tk := range list {
			if tk.Status == domain.TicketStatusWaiting && tk.OrderNumber < maxReceived {
				t.Errorf("屋台 %s の %d 番は waiting なのに、後ろの %d 番が received です", s.StallNumber, tk.OrderNumber, maxReceived)
			}
			if tk.Status == domain.TicketStatusReady && tk.OrderNumber < maxReceived {
				stale = true
			}
		}
		if stale {
			staleStalls++
		}

		switch s.BusinessStatus {
		case domain.BusinessStatusOpen:
			switch {
			case waiting >= 10:
				colors["red"]++
			case waiting >= 5:
				colors["yellow"]++
			default:
				colors["green"]++
			}
		case domain.BusinessStatusPreparing, domain.BusinessStatusClosed:
			if waiting != 0 {
				t.Errorf("屋台 %s は %s なのに waiting が %d 件あります", s.StallNumber, s.BusinessStatus, waiting)
			}
		}
	}
	for _, c := range []string{"red", "yellow", "green"} {
		if colors[c] == 0 {
			t.Errorf("混雑度が %s の営業中の屋台がありません（%v）", c, colors)
		}
	}
	if staleStalls == 0 {
		t.Error("取りに来ていない古い ready がある屋台がありません")
	}
}

func TestGenerateTickets_SameSeedSameData(t *testing.T) {
	stalls := testdataStalls(t)
	a, err := GenerateTickets(stalls, testTicketCount, 42)
	if err != nil {
		t.Fatal(err)
	}
	// 渡す順番が違っても同じデータになる。
	reversed := make([]domain.Stall, len(stalls))
	for i, s := range stalls {
		reversed[len(stalls)-1-i] = s
	}
	b, err := GenerateTickets(reversed, testTicketCount, 42)
	if err != nil {
		t.Fatal(err)
	}
	for i := range a {
		if a[i].PublicToken != b[i].PublicToken || a[i].StallID != b[i].StallID || a[i].Status != b[i].Status || !a[i].CreatedAt.Equal(b[i].CreatedAt) {
			t.Fatalf("%d 件目が違います: %+v と %+v", i+1, a[i], b[i])
		}
	}

	c, err := GenerateTickets(stalls, testTicketCount, 43)
	if err != nil {
		t.Fatal(err)
	}
	if a[0].PublicToken == c[0].PublicToken {
		t.Error("seed を変えても public_token が同じです")
	}
}

func TestGenerateTickets_Errors(t *testing.T) {
	stalls := testdataStalls(t)
	if _, err := GenerateTickets(stalls, 10, 42); err == nil {
		t.Error("80店に10件では足りないので、エラーになるはずです")
	}
	if _, err := GenerateTickets(nil, 100, 42); err == nil {
		t.Error("屋台がないので、エラーになるはずです")
	}
	noID := []domain.Stall{{StallNumber: "A-01", BusinessStatus: domain.BusinessStatusOpen}}
	if _, err := GenerateTickets(noID, 100, 42); err == nil {
		t.Error("屋台に ID がないので、エラーになるはずです")
	}
}

// DB に入れて、件数・集計・入れ直しを確かめる。
func TestSeedTickets(t *testing.T) {
	gdb := openMigrated(t)
	f, err := os.Open(filepath.Join("..", "..", "..", "testdata", "stalls.csv"))
	if err != nil {
		t.Fatal(err)
	}
	csvStalls, err := ReadStallsCSV(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err := SeedStalls(gdb, csvStalls); err != nil {
		t.Fatal(err)
	}
	stalls, err := ListStalls(gdb)
	if err != nil {
		t.Fatal(err)
	}
	tickets, err := GenerateTickets(stalls, testTicketCount, 42)
	if err != nil {
		t.Fatal(err)
	}

	var progressed int
	if err := SeedTickets(gdb, tickets, false, func(done int) { progressed = done }); err != nil {
		t.Fatal(err)
	}
	if progressed != testTicketCount {
		t.Errorf("progress に最後に渡された件数が %d です（%d のはず）", progressed, testTicketCount)
	}

	sum, err := SummarizeTickets(gdb)
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, n := range sum.ByStatus {
		total += n
	}
	if total != testTicketCount || len(sum.Stalls) != 80 {
		t.Errorf("整理券 %d 件・屋台 %d 店です（%d 件・80店のはず）", total, len(sum.Stalls), testTicketCount)
	}
	var waiting int
	for _, s := range sum.Stalls {
		waiting += s.Waiting
	}
	if int64(waiting) != sum.ByStatus[domain.TicketStatusWaiting] {
		t.Errorf("屋台ごとの待ち人数の合計 %d が、waiting の件数 %d と合いません", waiting, sum.ByStatus[domain.TicketStatusWaiting])
	}

	// 整理券があるときは、reset なしだとエラーにし、reset ありなら入れ直す。
	if err := SeedTickets(gdb, tickets, false, nil); err == nil {
		t.Error("整理券がすでにあるので、エラーになるはずです")
	}
	if err := SeedTickets(gdb, tickets, true, nil); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := gdb.Model(&domain.Ticket{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != testTicketCount {
		t.Errorf("入れ直したあと整理券が %d 件です（%d 件のはず）", n, testTicketCount)
	}
}
