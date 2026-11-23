package notify

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// newSubscription は、ブラウザが作るのと同じ形の鍵を持つ購読を作る。
func newSubscription(t *testing.T, endpoint string) domain.PushSubscription {
	t.Helper()
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	auth := make([]byte, 16)
	_, _ = rand.Read(auth)
	return domain.PushSubscription{
		Endpoint: endpoint,
		P256dh:   base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()),
		Auth:     base64.RawURLEncoding.EncodeToString(auth),
	}
}

func newTestWebPush(t *testing.T, subject string) *WebPush {
	t.Helper()
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	return NewWebPush(pub, priv, subject)
}

func TestSend(t *testing.T) {
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	w := newTestWebPush(t, "mailto:atokuru@example.com")
	if err := w.Send(t.Context(), newSubscription(t, srv.URL+"/push"), []byte(`{"type":"ready"}`)); err != nil {
		t.Fatal(err)
	}
	if got.Header.Get("Content-Encoding") != "aes128gcm" || got.Header.Get("TTL") != "1800" || got.Header.Get("Urgency") != "high" {
		t.Errorf("ヘッダー = %v", got.Header)
	}

	// VAPID の JWT の sub が mailto:mailto:... になっていないか。
	authz := got.Header.Get("Authorization")
	token := strings.TrimPrefix(strings.Split(authz, ",")[0], "vapid t=")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("Authorization = %q", authz)
	}
	claims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var c struct{ Sub string }
	if err := json.Unmarshal(claims, &c); err != nil {
		t.Fatal(err)
	}
	if c.Sub != "mailto:atokuru@example.com" {
		t.Errorf("sub = %q, want mailto:atokuru@example.com", c.Sub)
	}
}

func TestSend_StatusCodes(t *testing.T) {
	tests := []struct {
		status   int
		wantGone bool
		wantErr  bool
	}{
		{http.StatusCreated, false, false},
		{http.StatusNotFound, true, true},
		{http.StatusGone, true, true},
		{http.StatusTooManyRequests, false, true},
		{http.StatusInternalServerError, false, true},
		{http.StatusFound, false, true}, // リダイレクトには付いていかない
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			redirected := false
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/elsewhere" {
					redirected = true
				}
				if tt.status == http.StatusFound {
					w.Header().Set("Location", "/elsewhere")
				}
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			err := newTestWebPush(t, "https://example.com").Send(t.Context(), newSubscription(t, srv.URL+"/push"), []byte(`{}`))
			if gone := errors.Is(err, domain.ErrPushSubscriptionGone); gone != tt.wantGone {
				t.Errorf("gone = %v, want %v (err: %v)", gone, tt.wantGone, err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if redirected {
				t.Error("リダイレクト先にリクエストを送っています")
			}
		})
	}
}
