package handler

import (
	"net/http"
	"testing"
)

// まだ中身のない API が、ルーティングには登録されていて 501 を返すことを確かめる。
// API を実装したら、その行を消す。
func TestNotImplementedRoutes(t *testing.T) {
	e := newTestServer(New())
	routes := []struct{ method, path, body string }{
		{http.MethodGet, "/api/stalls", ""},
		{http.MethodGet, "/api/stalls/1", ""},
		{http.MethodPatch, "/api/stalls/1/business-status", `{"business_status":"open"}`},
		{http.MethodGet, "/api/stalls/1/tickets?status=waiting", ""},
		{http.MethodPost, "/api/stalls/1/tickets", `{}`},
		{http.MethodGet, "/api/tickets/abc", ""},
		{http.MethodPatch, "/api/tickets/1/status", `{"status":"ready"}`},
	}
	for _, r := range routes {
		if rec := doRequest(e, r.method, r.path, r.body); rec.Code != http.StatusNotImplemented {
			t.Errorf("%s %s: status = %d, want 501 (body: %s)", r.method, r.path, rec.Code, rec.Body)
		}
	}
}
