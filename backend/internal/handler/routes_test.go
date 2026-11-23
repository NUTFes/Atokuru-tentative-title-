package handler

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/infra/db"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/repository"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/usecase"
)

// testVapidPublicKey はテスト用の VAPID 公開鍵（中身は使わないので、形だけ合わせた値）。
const testVapidPublicKey = "BTestPublicKey"

// newTestServer は、手元の SQLite ファイルを使って main.go と同じ組み立てのサーバーを作る。
func newTestServer(t *testing.T) (*echo.Echo, *gorm.DB) {
	t.Helper()
	gdb, err := db.Open("file:"+filepath.Join(t.TempDir(), "test.db"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(gdb); err != nil {
		t.Fatal(err)
	}
	stallRepo := repository.NewStallRepository(gdb)
	ticketRepo := repository.NewTicketRepository(gdb)
	pushRepo := repository.NewPushSubscriptionRepository(gdb)
	h := New(
		usecase.NewStallUsecase(stallRepo),
		usecase.NewTicketUsecase(stallRepo, ticketRepo, pushRepo, nil),
		usecase.NewPushUsecase(ticketRepo, pushRepo, testVapidPublicKey),
	)

	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	gen.RegisterHandlersWithBaseURL(e, gen.NewStrictHandler(h, nil), "/api")
	return e, gdb
}

func doRequest(e *echo.Echo, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	return serve(e, req)
}

func newRequest(method, path string) *http.Request {
	return httptest.NewRequest(method, path, nil)
}

func serve(e *echo.Echo, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// 仕様書の全 API がルーティングに登録されているか（404 にならないか）と、
// エラーが Error の形で返るかを確かめる。
func TestRoutes(t *testing.T) {
	e, _ := newTestServer(t)

	tests := []struct {
		name, method, path, body string
		want                     int
		wantCode                 string
	}{
		{"存在しない URL", http.MethodGet, "/api/nope", "", http.StatusNotFound, "not_found"},
		{"stallId が数字でない", http.MethodGet, "/api/stalls/abc", "", http.StatusBadRequest, "bad_request"},
		{"JSON が壊れている", http.MethodPost, "/api/stalls/1/tickets", "{", http.StatusBadRequest, "bad_request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(e, tt.method, tt.path, tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Errorf("body = %s, want code %q", rec.Body, tt.wantCode)
			}
		})
	}

	if rec := doRequest(e, http.MethodGet, "/api/health", ""); rec.Code != http.StatusOK {
		t.Errorf("/api/health: status = %d", rec.Code)
	}
}

func itoa(id uint) string { return strconv.FormatUint(uint64(id), 10) }
