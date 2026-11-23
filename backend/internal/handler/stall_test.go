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
