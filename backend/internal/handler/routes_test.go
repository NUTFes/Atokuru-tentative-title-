package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

func newTestServer(h *Handler) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = HTTPErrorHandler
	gen.RegisterHandlersWithBaseURL(e, gen.NewStrictHandler(h, nil), "/api")
	return e
}

func doRequest(e *echo.Echo, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// 仕様書の全 API がルーティングに登録されているか（404 にならないか）と、
// エラーが Error の形で返るかを確かめる。
func TestRoutes(t *testing.T) {
	e := newTestServer(New())

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
