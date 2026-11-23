package translate

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestOpenAI は、OpenAI の代わりに handler が応答する手元のサーバーに向けた翻訳を作る。
func newTestOpenAI(t *testing.T, handler http.HandlerFunc) *OpenAI {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	o := NewOpenAI("sk-test", "test-model")
	o.baseURL = srv.URL
	return o
}

// chatResponse は Chat Completions の応答の形にする。
func chatResponse(content string) string {
	b, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{"message": map[string]string{"role": "assistant", "content": content}}},
	})
	return string(b)
}

func TestTranslate(t *testing.T) {
	var gotReq struct {
		Model          string              `json:"model"`
		Messages       []map[string]string `json:"messages"`
		ResponseFormat map[string]string   `json:"response_format"`
	}
	var gotAuth, gotPath string
	o := newTestOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath = r.Header.Get("Authorization"), r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotReq)
		// 頼んでいない言語は捨てる。
		_, _ = w.Write([]byte(chatResponse(`{"en": "Yakisoba, large", "fr": "Yakisoba, grand"}`)))
	})

	got, err := o.Translate(t.Context(), "焼きそば 大盛り", []string{"en"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got["en"] != "Yakisoba, large" {
		t.Errorf("got %v", got)
	}
	if gotAuth != "Bearer sk-test" || gotPath != "/chat/completions" {
		t.Errorf("Authorization = %q, path = %q", gotAuth, gotPath)
	}
	if gotReq.Model != "test-model" || gotReq.ResponseFormat["type"] != "json_object" || len(gotReq.Messages) != 2 {
		t.Errorf("リクエスト = %+v", gotReq)
	}
	if !strings.Contains(gotReq.Messages[1]["content"], "焼きそば 大盛り") || !strings.Contains(gotReq.Messages[1]["content"], `"en"`) {
		t.Errorf("訳す文と言語が渡っていません: %q", gotReq.Messages[1]["content"])
	}
}

func TestTranslate_Errors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"API がエラー", http.StatusUnauthorized, `{"error":{"message":"invalid api key"}}`},
		{"応答が JSON でない", http.StatusOK, `not json`},
		{"訳がない", http.StatusOK, `{"choices":[]}`},
		{"訳が JSON でない", http.StatusOK, chatResponse("Yakisoba, large")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := newTestOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})
			if got, err := o.Translate(t.Context(), "大盛り", []string{"en"}); err == nil {
				t.Errorf("エラーになりません: %v", got)
			}
		})
	}
}
