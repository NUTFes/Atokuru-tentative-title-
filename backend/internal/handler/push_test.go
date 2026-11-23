package handler

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/usecase"
)

// testP256dh と testAuth は、形（バイト数）だけ正しいテスト用の鍵。
var (
	testP256dh = base64.RawURLEncoding.EncodeToString(make([]byte, 65))
	testAuth   = base64.RawURLEncoding.EncodeToString(make([]byte, 16))
)

// subscriptionJSON は、ブラウザの PushSubscription.toJSON() と同じ形の JSON を作る。
func subscriptionJSON(endpoint, p256dh, auth string) string {
	return `{"endpoint":"` + endpoint + `","expirationTime":null,"keys":{"p256dh":"` + p256dh + `","auth":"` + auth + `"}}`
}

func TestGetVapidPublicKey(t *testing.T) {
	e, _ := newTestServer(t)

	rec := doRequest(e, http.MethodGet, "/api/push/vapid-public-key", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if got := decode[gen.VapidPublicKey](t, rec.Body.Bytes()); got.PublicKey != testVapidPublicKey {
		t.Errorf("public_key = %q, want %q", got.PublicKey, testVapidPublicKey)
	}
}

// 鍵が設定されていないときは 503（push_unavailable）を返す。
func TestGetVapidPublicKey_Unavailable(t *testing.T) {
	h := New(nil, nil, usecase.NewPushUsecase(nil, nil, ""))
	_, err := h.GetVapidPublicKey(t.Context(), gen.GetVapidPublicKeyRequestObject{})
	status, body := toErrorResponse(err)
	if status != http.StatusServiceUnavailable || body.Code != "push_unavailable" {
		t.Errorf("status = %d, code = %q, want 503 push_unavailable", status, body.Code)
	}
}

func TestCreatePushSubscription(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	ticket := issueTicket(t, e, s.ID, `{}`)
	path := "/api/tickets/" + ticket.PublicToken + "/push-subscriptions"
	endpoint := "https://fcm.googleapis.com/fcm/send/abc123"

	rec := doRequest(e, http.MethodPost, path, subscriptionJSON(endpoint, testP256dh, testAuth))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}

	// 同じ endpoint を新しい鍵でもう一度登録すると、1件のまま鍵が書き換わる。
	newAuth := base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef"))
	rec = doRequest(e, http.MethodPost, path, subscriptionJSON(endpoint, testP256dh, newAuth))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("2回目: status = %d, body = %s", rec.Code, rec.Body)
	}
	var subs []domain.PushSubscription
	if err := gdb.Find(&subs).Error; err != nil {
		t.Fatal(err)
	}
	if len(subs) != 1 {
		t.Fatalf("購読が %d 件あります（1件のはず）", len(subs))
	}
	if subs[0].TicketID != uint(ticket.Id) || subs[0].Endpoint != endpoint || subs[0].Auth != newAuth {
		t.Errorf("購読 = %+v", subs[0])
	}

	// 別の端末（endpoint）なら、同じ整理券に2件目として入る。
	rec = doRequest(e, http.MethodPost, path, subscriptionJSON("https://web.push.apple.com/xyz", testP256dh, testAuth))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("別の端末: status = %d, body = %s", rec.Code, rec.Body)
	}
	var n int64
	gdb.Model(&domain.PushSubscription{}).Count(&n)
	if n != 2 {
		t.Errorf("購読が %d 件あります（2件のはず）", n)
	}
}

func TestCreatePushSubscription_Errors(t *testing.T) {
	e, gdb := newTestServer(t)
	s := createStall(t, gdb, "A-01", domain.BusinessStatusOpen)
	ticket := issueTicket(t, e, s.ID, `{}`)
	received := issueTicket(t, e, s.ID, `{}`)
	if rec := doRequest(e, http.MethodPatch, "/api/tickets/"+itoa(uint(received.Id))+"/status", `{"status":"received"}`); rec.Code != http.StatusOK {
		t.Fatalf("受け取り済みにできません: %s", rec.Body)
	}
	ok := "https://fcm.googleapis.com/fcm/send/abc123"

	tests := []struct {
		name, token, body string
		want              int
		wantCode          string
	}{
		{"整理券がない", "nope", subscriptionJSON(ok, testP256dh, testAuth), http.StatusNotFound, "not_found"},
		{"受け取り済み", received.PublicToken, subscriptionJSON(ok, testP256dh, testAuth), http.StatusBadRequest, "bad_request"},
		{"http の URL", ticket.PublicToken, subscriptionJSON("http://fcm.googleapis.com/x", testP256dh, testAuth), http.StatusBadRequest, "bad_request"},
		{"IP アドレスの URL", ticket.PublicToken, subscriptionJSON("https://169.254.169.254/x", testP256dh, testAuth), http.StatusBadRequest, "bad_request"},
		{"localhost の URL", ticket.PublicToken, subscriptionJSON("https://localhost:8080/x", testP256dh, testAuth), http.StatusBadRequest, "bad_request"},
		{"endpoint が長すぎる", ticket.PublicToken, subscriptionJSON("https://fcm.googleapis.com/"+strings.Repeat("a", 2048), testP256dh, testAuth), http.StatusBadRequest, "bad_request"},
		{"p256dh の長さが違う", ticket.PublicToken, subscriptionJSON(ok, testAuth, testAuth), http.StatusBadRequest, "bad_request"},
		{"auth が base64url でない", ticket.PublicToken, subscriptionJSON(ok, testP256dh, "!!!"), http.StatusBadRequest, "bad_request"},
		{"keys がない", ticket.PublicToken, `{"endpoint":"` + ok + `"}`, http.StatusBadRequest, "bad_request"},
		{"JSON が壊れている", ticket.PublicToken, `{`, http.StatusBadRequest, "bad_request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(e, http.MethodPost, "/api/tickets/"+tt.token+"/push-subscriptions", tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Errorf("body = %s, want code %q", rec.Body, tt.wantCode)
			}
		})
	}

	var n int64
	gdb.Model(&domain.PushSubscription{}).Count(&n)
	if n != 0 {
		t.Errorf("エラーのときに購読が %d 件入っています", n)
	}
}
