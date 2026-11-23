package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/usecase"
)

// postTemplate は API で定型文を登録する。
func postTemplate(e *echo.Echo, stallID uint, comment string) *httptest.ResponseRecorder {
	return doRequest(e, http.MethodPost, "/api/stalls/"+itoa(stallID)+"/comment-templates", `{"comment":"`+comment+`"}`)
}

func TestCreateCommentTemplate(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)

	rec := postTemplate(e, s.ID, "  焼きそば 大盛り  ")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	got := decode[gen.CommentTemplate](t, rec.Body.Bytes())
	if got.StallId != int64(s.ID) || got.Comment != "焼きそば 大盛り" || got.Translations["en"] != "EN:焼きそば 大盛り" {
		t.Errorf("got %+v", got)
	}

	// 一覧は登録した順。
	postTemplate(e, s.ID, "ソース抜き")
	rec = doRequest(e, http.MethodGet, "/api/stalls/"+itoa(s.ID)+"/comment-templates", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("一覧: status = %d, body = %s", rec.Code, rec.Body)
	}
	list := decode[[]gen.CommentTemplate](t, rec.Body.Bytes())
	if len(list) != 2 || list[0].Comment != "焼きそば 大盛り" || list[1].Comment != "ソース抜き" || list[1].Translations["en"] != "EN:ソース抜き" {
		t.Errorf("一覧 = %+v", list)
	}

	// ほかの屋台の一覧には出ない。
	other := createStall(t, gdb, "A-02", domain.BusinessStatusOpen)
	rec = doRequest(e, http.MethodGet, "/api/stalls/"+itoa(other.ID)+"/comment-templates", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("ほかの屋台の一覧: %d %s", rec.Code, rec.Body)
	}
}

func TestCreateCommentTemplate_Errors(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	if rec := postTemplate(e, s.ID, "焼きそば 大盛り"); rec.Code != http.StatusCreated {
		t.Fatalf("登録できません: %s", rec.Body)
	}

	tests := []struct {
		name     string
		stallID  uint
		body     string
		want     int
		wantCode string
	}{
		{"同じ文", s.ID, `{"comment":"焼きそば 大盛り"}`, http.StatusConflict, "conflict"},
		{"前後の空白だけ違う同じ文", s.ID, `{"comment":" 焼きそば 大盛り "}`, http.StatusConflict, "conflict"},
		{"空", s.ID, `{"comment":"   "}`, http.StatusBadRequest, "bad_request"},
		{"長すぎる", s.ID, `{"comment":"` + strings.Repeat("あ", 201) + `"}`, http.StatusBadRequest, "bad_request"},
		{"屋台がない", 9999, `{"comment":"大盛り"}`, http.StatusNotFound, "not_found"},
		{"JSON が壊れている", s.ID, `{`, http.StatusBadRequest, "bad_request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(e, http.MethodPost, "/api/stalls/"+itoa(tt.stallID)+"/comment-templates", tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Errorf("body = %s, want code %q", rec.Body, tt.wantCode)
			}
		})
	}

	// 一覧で屋台がないときは 404。
	if rec := doRequest(e, http.MethodGet, "/api/stalls/9999/comment-templates", ""); rec.Code != http.StatusNotFound {
		t.Errorf("一覧で屋台がない: status = %d", rec.Code)
	}
}

func TestCreateCommentTemplate_Limit(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	for i := 1; i <= 30; i++ {
		if rec := postTemplate(e, s.ID, "定型文"+strconv.Itoa(i)); rec.Code != http.StatusCreated {
			t.Fatalf("%d 件目: status = %d, body = %s", i, rec.Code, rec.Body)
		}
	}
	rec := postTemplate(e, s.ID, "31件目")
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"conflict"`) {
		t.Errorf("31 件目: status = %d, body = %s", rec.Code, rec.Body)
	}
}

// 翻訳の設定がないときや、訳すのに失敗したときも、訳なし（{}）で登録できる。
func TestCreateCommentTemplate_WithoutTranslation(t *testing.T) {
	tests := []struct {
		name       string
		translator usecase.Translator
	}{
		{"設定がない", nil},
		{"失敗する", failingTranslator{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, gdb := newTestServerWithTranslator(t, tt.translator)
			s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
			rec := postTemplate(e, s.ID, "大盛り")
			if rec.Code != http.StatusCreated {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"translations":{}`) {
				t.Errorf("body = %s, want translations {}", rec.Body)
			}
		})
	}
}

func TestDeleteCommentTemplate(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	tpl := decode[gen.CommentTemplate](t, postTemplate(e, s.ID, "大盛り").Body.Bytes())
	path := "/api/comment-templates/" + strconv.FormatInt(tpl.Id, 10)

	if rec := doRequest(e, http.MethodDelete, path, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if rec := doRequest(e, http.MethodDelete, path, ""); rec.Code != http.StatusNotFound {
		t.Errorf("2回目: status = %d, want 404", rec.Code)
	}
	// 消したあとは同じ文をもう一度登録できる。
	if rec := postTemplate(e, s.ID, "大盛り"); rec.Code != http.StatusCreated {
		t.Errorf("消したあとの登録: status = %d, body = %s", rec.Code, rec.Body)
	}
}

// 定型文と同じ備考で発券すると、整理券に訳が入る。自由入力の備考や、備考なしでは {}。
func TestGetTicket_CommentTranslations(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	other := createStall(t, gdb, "A-02", domain.BusinessStatusOpen)
	postTemplate(e, s.ID, "焼きそば 大盛り")

	tests := []struct {
		name    string
		stallID uint
		body    string
		wantEn  string
	}{
		{"定型文と同じ", s.ID, `{"comment":"焼きそば 大盛り"}`, "EN:焼きそば 大盛り"},
		{"前後の空白だけ違う", s.ID, `{"comment":"  焼きそば 大盛り "}`, "EN:焼きそば 大盛り"},
		{"自由入力", s.ID, `{"comment":"焼きそば 大盛り 青のり抜き"}`, ""},
		{"備考なし", s.ID, `{}`, ""},
		{"ほかの屋台の定型文", other.ID, `{"comment":"焼きそば 大盛り"}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := issueTicket(t, e, tt.stallID, tt.body)
			rec := doRequest(e, http.MethodGet, "/api/tickets/"+ticket.PublicToken, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
			}
			got := decode[gen.TicketDetail](t, rec.Body.Bytes())
			if got.CommentTranslations == nil {
				t.Fatalf("comment_translations が null です: %s", rec.Body)
			}
			if got.CommentTranslations["en"] != tt.wantEn {
				t.Errorf("comment_translations = %v, want en %q", got.CommentTranslations, tt.wantEn)
			}
		})
	}

	// 定型文を消しても、発券済みの整理券の訳は残る。
	ticket := issueTicket(t, e, s.ID, `{"comment":"焼きそば 大盛り"}`)
	list := decode[[]gen.CommentTemplate](t, doRequest(e, http.MethodGet, "/api/stalls/"+itoa(s.ID)+"/comment-templates", "").Body.Bytes())
	doRequest(e, http.MethodDelete, "/api/comment-templates/"+strconv.FormatInt(list[0].Id, 10), "")
	got := decode[gen.TicketDetail](t, doRequest(e, http.MethodGet, "/api/tickets/"+ticket.PublicToken, "").Body.Bytes())
	if got.CommentTranslations["en"] != "EN:焼きそば 大盛り" {
		t.Errorf("定型文を消したあとの訳 = %v", got.CommentTranslations)
	}
}

type failingTranslator struct{}

func (failingTranslator) Translate(context.Context, string, []string) (domain.Translations, error) {
	return nil, errors.New("OpenAI が落ちている")
}
