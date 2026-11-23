package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

func ptr[T any](v T) *T { return &v }

// createStall はテスト用の屋台を DB に入れる。
func createStall(t *testing.T, gdb *gorm.DB, number string, status domain.BusinessStatus) *domain.Stall {
	t.Helper()
	pw := "secret-hash"
	s := &domain.Stall{
		Name:           "屋台" + number,
		StallNumber:    number,
		PasswordHash:   &pw,
		HPLink:         ptr("https://example.com/" + number),
		BusinessStatus: status,
	}
	if err := gdb.Create(s).Error; err != nil {
		t.Fatal(err)
	}
	return s
}

func TestGetStall(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)

	rec := doRequest(e, http.MethodGet, "/api/stalls/"+itoa(s.ID), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "password") || strings.Contains(rec.Body.String(), "secret-hash") {
		t.Errorf("パスワードが返っています: %s", rec.Body)
	}
	var got gen.Stall
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := gen.Stall{
		Id: int64(s.ID), Name: "屋台A-01", StallNumber: "A-01",
		HpLink: ptr("https://example.com/A-01"), BusinessStatus: gen.Open,
	}
	if got.Id != want.Id || got.Name != want.Name || got.StallNumber != want.StallNumber ||
		*got.HpLink != *want.HpLink || got.BusinessStatus != want.BusinessStatus {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGetStall_NotFound(t *testing.T) {
	e, _ := newTestServer(t)
	for _, path := range []string{"/api/stalls/999", "/api/stalls/0", "/api/stalls/-1"} {
		rec := doRequest(e, http.MethodGet, path, "")
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
			t.Errorf("%s: status = %d, body = %s", path, rec.Code, rec.Body)
		}
	}
}

func TestUpdateBusinessStatus(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusPreparing)
	path := "/api/stalls/" + itoa(s.ID) + "/business-status"

	for _, status := range []string{"open", "sold_out", "closed", "preparing", "open"} {
		rec := doRequest(e, http.MethodPatch, path, `{"business_status":"`+status+`"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, body = %s", status, rec.Code, rec.Body)
		}
		if got := decode[gen.Stall](t, rec.Body.Bytes()); string(got.BusinessStatus) != status {
			t.Errorf("got %s, want %s", got.BusinessStatus, status)
		}
		// GET でも変わっている
		got := decode[gen.Stall](t, doRequest(e, http.MethodGet, "/api/stalls/"+itoa(s.ID), "").Body.Bytes())
		if string(got.BusinessStatus) != status {
			t.Errorf("GET: got %s, want %s", got.BusinessStatus, status)
		}
	}

	// 営業中にしたら発券でき、閉店にしたら発券できない
	issueTicket(t, e, s.ID, `{}`)
	doRequest(e, http.MethodPatch, path, `{"business_status":"closed"}`)
	if rec := doRequest(e, http.MethodPost, "/api/stalls/"+itoa(s.ID)+"/tickets", `{}`); rec.Code != http.StatusConflict {
		t.Errorf("閉店で発券: status = %d", rec.Code)
	}
}

func TestUpdateBusinessStatus_Errors(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	path := "/api/stalls/" + itoa(s.ID) + "/business-status"

	tests := []struct {
		path, body string
		want       int
		wantCode   string
	}{
		{path, `{"business_status":"busy"}`, http.StatusBadRequest, "bad_request"},
		{path, `{}`, http.StatusBadRequest, "bad_request"},
		{"/api/stalls/999/business-status", `{"business_status":"open"}`, http.StatusNotFound, "not_found"},
	}
	for _, tt := range tests {
		rec := doRequest(e, http.MethodPatch, tt.path, tt.body)
		if rec.Code != tt.want || !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
			t.Errorf("%s %s: status = %d, body = %s", tt.path, tt.body, rec.Code, rec.Body)
		}
	}
}

func TestListStalls(t *testing.T) {
	e, gdb := newTestServer(t)
	// 店番号の順に並ぶことを確かめるため、わざと逆の順に作る
	busy := createStall(t, gdb, "B-01", domain.BusinessStatusOpen)
	closed := createStall(t, gdb, "A-02", domain.BusinessStatusClosed)
	empty := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	empty.HPLink = nil
	if err := gdb.Model(empty).Update("hp_link", nil).Error; err != nil {
		t.Fatal(err)
	}

	// busy に12枚発行し、うち2枚を ready にする → waiting は10枚
	for i := 0; i < 12; i++ {
		tk := issueTicket(t, e, busy.ID, `{}`)
		if i < 2 {
			doRequest(e, http.MethodPatch, "/api/tickets/"+itoa(uint(tk.Id))+"/status", `{"status":"ready"}`)
		}
	}
	// closed には営業中のときに3枚発行してから閉店にする
	if err := gdb.Model(closed).Update("business_status", domain.BusinessStatusOpen).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		issueTicket(t, e, closed.ID, `{}`)
	}
	doRequest(e, http.MethodPatch, "/api/stalls/"+itoa(closed.ID)+"/business-status", `{"business_status":"closed"}`)

	rec := doRequest(e, http.MethodGet, "/api/stalls", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	got := decode[[]gen.StallSummary](t, rec.Body.Bytes())
	if len(got) != 3 {
		t.Fatalf("got %d stalls: %s", len(got), rec.Body)
	}

	want := []struct {
		number     string
		waiting    int
		congestion gen.Congestion
		hpLink     bool
	}{
		{"A-01", 0, gen.Green, false},
		{"A-02", 3, gen.Gray, true},
		{"B-01", 10, gen.Red, true},
	}
	for i, w := range want {
		g := got[i]
		if g.StallNumber != w.number || g.WaitingCount != w.waiting || g.Congestion != w.congestion || (g.HpLink != nil) != w.hpLink {
			t.Errorf("%d 番目: got %+v, want %+v", i, g, w)
		}
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Errorf("パスワードが返っています: %s", rec.Body)
	}
}

func TestListStalls_Empty(t *testing.T) {
	e, _ := newTestServer(t)
	rec := doRequest(e, http.MethodGet, "/api/stalls", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("status = %d, body = %s", rec.Code, rec.Body)
	}
}
