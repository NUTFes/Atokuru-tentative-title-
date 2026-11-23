// Package notify は通知（Web Push）をプッシュサーバーに送る。
package notify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// ttlSeconds は、端末の電源が切れているときなどに、プッシュサーバーが通知を預かっておく時間。
// 調理完了やあと3人は、時間がたつと意味がなくなるので短めにする。
const ttlSeconds = 30 * 60

// WebPush は VAPID の鍵で署名して、Web Push の通知を送る。
type WebPush struct {
	publicKey  string
	privateKey string
	subscriber string
	client     *http.Client
}

// NewWebPush は送信の準備をする。subject は VAPID_SUBJECT（mailto: か https: の URL）。
func NewWebPush(publicKey, privateKey, subject string) *WebPush {
	return &WebPush{
		publicKey:  publicKey,
		privateKey: privateKey,
		// webpush-go は https: で始まらない値の前に mailto: を付けるので、付いていれば外しておく
		// （そのまま渡すと mailto:mailto:... になる）。
		subscriber: strings.TrimPrefix(subject, "mailto:"),
		client: &http.Client{
			Timeout: 10 * time.Second,
			// 購読の endpoint は来場者が送ってくる値なので、リダイレクトで別のサーバーに飛ばされないようにする。
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

// Send は購読1件に通知を送る。プッシュサーバーが「もう使えない購読」（404・410）と返したときは
// domain.ErrPushSubscriptionGone を返す。
func (w *WebPush) Send(ctx context.Context, sub domain.PushSubscription, payload []byte) error {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &webpush.Options{
		HTTPClient:      w.client,
		Subscriber:      w.subscriber,
		VAPIDPublicKey:  w.publicKey,
		VAPIDPrivateKey: w.privateKey,
		TTL:             ttlSeconds,
		Urgency:         webpush.UrgencyHigh,
	})
	if err != nil {
		return fmt.Errorf("通知を送れません: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return domain.ErrPushSubscriptionGone
	case resp.StatusCode >= 300:
		return errors.New("プッシュサーバーがエラーを返しました: " + resp.Status)
	}
	return nil
}
