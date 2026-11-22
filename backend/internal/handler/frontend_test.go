package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// out/ と同じ形のディレクトリを作る（trailingSlash: true のときの出力）。
func writeFrontend(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"index.html":                 "<h1>home</h1>",
		"map/index.html":             "<h1>map</h1>",
		"stall/tickets/index.html":   "<h1>stall tickets</h1>",
		"404.html":                   "<h1>not found page</h1>",
		"_next/static/chunks/app.js": "console.log(1)",
	}
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestServeFrontend(t *testing.T) {
	e, _ := newTestServer(t)
	ServeFrontend(e, writeFrontend(t))

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
		wantCache  string
	}{
		{"/", http.StatusOK, "<h1>home</h1>", "no-cache"},
		{"/?token=abc", http.StatusOK, "<h1>home</h1>", "no-cache"},
		{"/map/", http.StatusOK, "<h1>map</h1>", "no-cache"},
		{"/map", http.StatusOK, "<h1>map</h1>", "no-cache"},
		{"/stall/tickets/", http.StatusOK, "<h1>stall tickets</h1>", "no-cache"},
		{"/_next/static/chunks/app.js", http.StatusOK, "console.log(1)", "immutable"},
		{"/no-such-page/", http.StatusNotFound, "<h1>not found page</h1>", ""},
		// API は FE の配信に邪魔されない
		{"/api/health", http.StatusOK, `"status":"ok"`, "no-store"},
		{"/api/no-such-api", http.StatusNotFound, `"code":"not_found"`, ""},
	}
	for _, tt := range tests {
		rec := doRequest(e, http.MethodGet, tt.path, "")
		if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), tt.wantBody) {
			t.Errorf("%s: status = %d, body = %s", tt.path, rec.Code, rec.Body)
		}
		if tt.wantCache != "" && !strings.Contains(rec.Header().Get("Cache-Control"), tt.wantCache) {
			t.Errorf("%s: Cache-Control = %q, want %q", tt.path, rec.Header().Get("Cache-Control"), tt.wantCache)
		}
	}

	// ディレクトリの外のファイルは読めない
	if rec := doRequest(e, http.MethodGet, "/../go.mod", ""); rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "module") {
		t.Errorf("ディレクトリの外のファイルが読めてしまう: %s", rec.Body)
	}
}

func TestAllowCORS(t *testing.T) {
	e, _ := newTestServer(t)
	AllowCORS(e, []string{"http://localhost:3000"})

	// ブラウザが PATCH の前に送る確認（プリフライト）
	req := newRequest(http.MethodOptions, "/api/tickets/1/status")
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	rec := serve(e, req)
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" ||
		!strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), http.MethodPatch) {
		t.Errorf("許可したオリジン: status = %d, headers = %v", rec.Code, rec.Header())
	}

	// 許可していないオリジンには Allow-Origin を返さない
	req = newRequest(http.MethodGet, "/api/health")
	req.Header.Set("Origin", "https://evil.example.com")
	if rec := serve(e, req); rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("許可していないオリジンに Allow-Origin を返している: %v", rec.Header())
	}
}
