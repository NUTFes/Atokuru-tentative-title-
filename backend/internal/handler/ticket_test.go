package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

func decode[T any](t *testing.T, body []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("JSON を読めません: %v (%s)", err, body)
	}
	return v
}

// issueTicket は API で整理券を発行する。
func issueTicket(t *testing.T, e *echo.Echo, stallID uint, body string) gen.Ticket {
	t.Helper()
	rec := doRequest(e, http.MethodPost, "/api/stalls/"+itoa(stallID)+"/tickets", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("発券: status = %d, body = %s", rec.Code, rec.Body)
	}
	return decode[gen.Ticket](t, rec.Body.Bytes())
}

func TestCreateTicket(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	other := createStall(t, gdb, "A-02", domain.BusinessStatusOpen)

	t1 := issueTicket(t, e, s.ID, `{"comment":"  焼きそば 大盛り  "}`)
	if t1.OrderNumber != 1 || t1.Status != gen.Waiting || t1.StallId != int64(s.ID) {
		t.Errorf("1枚目: %+v", t1)
	}
	if t1.Comment == nil || *t1.Comment != "焼きそば 大盛り" {
		t.Errorf("備考の前後の空白が取れていません: %v", t1.Comment)
	}
	if len(t1.PublicToken) != 22 || t1.CalledAt != nil || t1.CreatedAt.IsZero() {
		t.Errorf("1枚目: %+v", t1)
	}

	t2 := issueTicket(t, e, s.ID, `{}`)
	if t2.OrderNumber != 2 || t2.Comment != nil || t2.PublicToken == t1.PublicToken {
		t.Errorf("2枚目: %+v", t2)
	}

	// 屋台ごとに 1 から振る
	if o := issueTicket(t, e, other.ID, `{"comment":""}`); o.OrderNumber != 1 || o.Comment != nil {
		t.Errorf("別の屋台: %+v", o)
	}
}

// 同時に発券しても order_number が重複しないことを確かめる。
func TestCreateTicket_Concurrent(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)

	const n = 30
	var wg sync.WaitGroup
	nums := make(chan int, n)
	errs := make(chan string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := doRequest(e, http.MethodPost, "/api/stalls/"+itoa(s.ID)+"/tickets", `{}`)
			if rec.Code != http.StatusCreated {
				errs <- rec.Body.String()
				return
			}
			var tk gen.Ticket
			_ = json.Unmarshal(rec.Body.Bytes(), &tk)
			nums <- tk.OrderNumber
		}()
	}
	wg.Wait()
	close(nums)
	close(errs)
	for msg := range errs {
		t.Errorf("発券に失敗: %s", msg)
	}
	seen := map[int]bool{}
	for num := range nums {
		if seen[num] {
			t.Errorf("order_number %d が重複しています", num)
		}
		seen[num] = true
	}
	for i := 1; i <= len(seen); i++ {
		if !seen[i] {
			t.Errorf("order_number %d が抜けています", i)
		}
	}
}

func TestCreateTicket_Errors(t *testing.T) {
	e, gdb := newTestServer(t)
	open := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)

	tests := []struct {
		name     string
		stallID  uint
		body     string
		want     int
		wantCode string
	}{
		{"存在しない屋台", 999, `{}`, http.StatusNotFound, "not_found"},
		{"備考が長すぎる", open.ID, `{"comment":"` + strings.Repeat("あ", 201) + `"}`, http.StatusBadRequest, "bad_request"},
	}
	for _, status := range []domain.BusinessStatus{domain.BusinessStatusPreparing, domain.BusinessStatusSoldOut, domain.BusinessStatusClosed} {
		s := createStall(t, gdb, "B-"+string(status), status)
		tests = append(tests, struct {
			name     string
			stallID  uint
			body     string
			want     int
			wantCode string
		}{"営業中でない（" + string(status) + "）", s.ID, `{}`, http.StatusConflict, "stall_not_open"})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(e, http.MethodPost, "/api/stalls/"+itoa(tt.stallID)+"/tickets", tt.body)
			if rec.Code != tt.want || !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Errorf("status = %d, body = %s", rec.Code, rec.Body)
			}
		})
	}

	// ちょうど 200 文字なら発券できる
	issueTicket(t, e, open.ID, `{"comment":"`+strings.Repeat("あ", 200)+`"}`)
}

func TestGetTicket(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)

	var issued []gen.Ticket
	for i := 0; i < 5; i++ {
		issued = append(issued, issueTicket(t, e, s.ID, `{"comment":"メモ"}`))
	}
	// 2番を受け取り済みにする（1番は取りに来ていない）
	if err := gdb.Model(&domain.Ticket{}).Where("id = ?", issued[1].Id).Update("status", domain.TicketStatusReceived).Error; err != nil {
		t.Fatal(err)
	}

	rec := doRequest(e, http.MethodGet, "/api/tickets/"+issued[4].PublicToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	got := decode[gen.TicketDetail](t, rec.Body.Bytes())
	// 5番 − 受け取り済みの最大 2番 − 1 = 2人（3番と4番）
	if got.PeopleAhead != 2 || got.OrderNumber != 5 || got.StallName != "屋台A-01" || got.StallNumber != "A-01" ||
		got.Status != gen.Waiting || got.Comment == nil || *got.Comment != "メモ" {
		t.Errorf("got %+v", got)
	}
	if strings.Contains(rec.Body.String(), `"id"`) {
		t.Errorf("連番の id を返しています: %s", rec.Body)
	}
}

func TestGetTicket_NotFound(t *testing.T) {
	e, _ := newTestServer(t)
	rec := doRequest(e, http.MethodGet, "/api/tickets/no-such-token", "")
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Errorf("status = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestListStallTickets(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	other := createStall(t, gdb, "A-02", domain.BusinessStatusOpen)
	issueTicket(t, e, other.ID, `{}`) // 別の屋台の整理券は混ざらない

	// 1〜6番を発行し、status を waiting, ready, received, waiting, received, ready にする
	statuses := []domain.TicketStatus{"waiting", "ready", "received", "waiting", "received", "ready"}
	for _, st := range statuses {
		tk := issueTicket(t, e, s.ID, `{}`)
		if err := gdb.Model(&domain.Ticket{}).Where("id = ?", tk.Id).Update("status", st).Error; err != nil {
			t.Fatal(err)
		}
	}

	orderNumbers := func(path string) []int {
		t.Helper()
		rec := doRequest(e, http.MethodGet, path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, body = %s", path, rec.Code, rec.Body)
		}
		var nums []int
		for _, tk := range decode[[]gen.Ticket](t, rec.Body.Bytes()) {
			if tk.StallId != int64(s.ID) {
				t.Errorf("別の屋台の整理券が混ざっています: %+v", tk)
			}
			nums = append(nums, tk.OrderNumber)
		}
		return nums
	}
	base := "/api/stalls/" + itoa(s.ID) + "/tickets"

	tests := []struct {
		query string
		want  []int
	}{
		{"?status=waiting&status=ready", []int{1, 2, 4, 6}},
		{"?status=received", []int{3, 5}},
		{"?status=received&order=desc", []int{5, 3}},
		{"?status=waiting&status=ready&limit=2", []int{1, 2}},
		{"?status=waiting&status=waiting", []int{1, 4}},
	}
	for _, tt := range tests {
		if got := orderNumbers(base + tt.query); fmt.Sprint(got) != fmt.Sprint(tt.want) {
			t.Errorf("%s: got %v, want %v", tt.query, got, tt.want)
		}
	}

	// 該当なしは空の配列（null ではない）
	empty := createStall(t, gdb, "A-03", domain.BusinessStatusOpen)
	if rec := doRequest(e, http.MethodGet, "/api/stalls/"+itoa(empty.ID)+"/tickets?status=waiting", ""); strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("空のとき: %s", rec.Body)
	}
}

func TestListStallTickets_Errors(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	base := "/api/stalls/" + itoa(s.ID) + "/tickets"

	tests := []struct {
		path     string
		want     int
		wantCode string
	}{
		{base, http.StatusBadRequest, "bad_request"},
		{base + "?status=cooking", http.StatusBadRequest, "bad_request"},
		{base + "?status=waiting&limit=0", http.StatusBadRequest, "bad_request"},
		{base + "?status=waiting&limit=501", http.StatusBadRequest, "bad_request"},
		{base + "?status=waiting&limit=abc", http.StatusBadRequest, "bad_request"},
		{base + "?status=waiting&order=random", http.StatusBadRequest, "bad_request"},
		{"/api/stalls/999/tickets?status=waiting", http.StatusNotFound, "not_found"},
	}
	for _, tt := range tests {
		rec := doRequest(e, http.MethodGet, tt.path, "")
		if rec.Code != tt.want || !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
			t.Errorf("%s: status = %d, body = %s", tt.path, rec.Code, rec.Body)
		}
	}
}

func TestUpdateTicketStatus(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	tk := issueTicket(t, e, s.ID, `{}`)
	path := "/api/tickets/" + strconv.FormatInt(tk.Id, 10) + "/status"

	patch := func(status string) gen.Ticket {
		t.Helper()
		rec := doRequest(e, http.MethodPatch, path, `{"status":"`+status+`"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, body = %s", status, rec.Code, rec.Body)
		}
		return decode[gen.Ticket](t, rec.Body.Bytes())
	}

	ready := patch("ready")
	if ready.Status != gen.Ready || ready.CalledAt == nil {
		t.Fatalf("ready にしたら called_at が入るはず: %+v", ready)
	}
	// 同じ status をもう一度送っても called_at は変わらない
	if again := patch("ready"); again.CalledAt == nil || !again.CalledAt.Equal(*ready.CalledAt) {
		t.Errorf("ready を2回押したら called_at が変わった: %v → %v", ready.CalledAt, again.CalledAt)
	}
	// 受け取り済みにしても called_at は残る
	if received := patch("received"); received.Status != gen.Received || received.CalledAt == nil {
		t.Errorf("received: %+v", received)
	}
	// 押し間違いを戻せる。waiting に戻したら called_at は空になる
	if back := patch("waiting"); back.Status != gen.Waiting || back.CalledAt != nil {
		t.Errorf("waiting に戻したとき: %+v", back)
	}
	// waiting から直接 received にもできる（called_at は空のまま）
	if direct := patch("received"); direct.Status != gen.Received || direct.CalledAt != nil {
		t.Errorf("waiting → received: %+v", direct)
	}
}

func TestUpdateTicketStatus_Errors(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	tk := issueTicket(t, e, s.ID, `{}`)
	path := "/api/tickets/" + strconv.FormatInt(tk.Id, 10) + "/status"

	tests := []struct {
		path, body string
		want       int
		wantCode   string
	}{
		{path, `{"status":"cooking"}`, http.StatusBadRequest, "bad_request"},
		{path, `{}`, http.StatusBadRequest, "bad_request"},
		{"/api/tickets/999/status", `{"status":"ready"}`, http.StatusNotFound, "not_found"},
	}
	for _, tt := range tests {
		rec := doRequest(e, http.MethodPatch, tt.path, tt.body)
		if rec.Code != tt.want || !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
			t.Errorf("%s %s: status = %d, body = %s", tt.path, tt.body, rec.Code, rec.Body)
		}
	}
}
