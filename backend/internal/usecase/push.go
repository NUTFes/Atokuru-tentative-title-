package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/repository"
)

// 購読の値の上限（仕様書の maxLength と合わせる）。
const (
	maxEndpointLength = 2048
	maxPushKeyLength  = 256
)

// 鍵を base64url から戻したときのバイト数（Web Push の決まり。RFC 8291）。
const (
	p256dhBytes = 65 // P-256 の公開鍵（非圧縮）
	authBytes   = 16
)

// PushUsecase は通知（Web Push）の購読についての処理をまとめる。
type PushUsecase struct {
	tickets       *repository.TicketRepository
	subscriptions *repository.PushSubscriptionRepository
	// publicKey は VAPID 公開鍵。空なら通知を使わない（鍵を設定していないローカル開発など）。
	publicKey string
}

func NewPushUsecase(tickets *repository.TicketRepository, subscriptions *repository.PushSubscriptionRepository, publicKey string) *PushUsecase {
	return &PushUsecase{tickets: tickets, subscriptions: subscriptions, publicKey: publicKey}
}

// VapidPublicKey は、ブラウザで購読を始めるときに使う VAPID 公開鍵を返す。
func (u *PushUsecase) VapidPublicKey() (string, error) {
	if u.publicKey == "" {
		return "", domain.PushUnavailable("通知は使えません")
	}
	return u.publicKey, nil
}

// SubscribeInput はブラウザの PushSubscription のうち、通知を送るのに必要なもの。
type SubscribeInput struct {
	Endpoint string
	P256dh   string
	Auth     string
}

// Subscribe は、public_token で指定した整理券に購読を登録する。
// 同じ整理券に同じ endpoint をもう一度登録したときは上書きする。受け取り済みの整理券には登録できない。
func (u *PushUsecase) Subscribe(ctx context.Context, publicToken string, in SubscribeInput) error {
	if err := validateEndpoint(in.Endpoint); err != nil {
		return err
	}
	if !validPushKey(in.P256dh, p256dhBytes) {
		return domain.InvalidInput("keys.p256dh が正しくありません")
	}
	if !validPushKey(in.Auth, authBytes) {
		return domain.InvalidInput("keys.auth が正しくありません")
	}

	t, err := u.tickets.FindByPublicToken(ctx, publicToken)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.NotFound("整理券が見つかりません")
	}
	if err != nil {
		return err
	}
	if t.Status == domain.TicketStatusReceived {
		return domain.InvalidInput("受け取り済みの整理券には通知を登録できません")
	}

	return u.subscriptions.Upsert(ctx, &domain.PushSubscription{
		TicketID: t.ID,
		Endpoint: in.Endpoint,
		P256dh:   in.P256dh,
		Auth:     in.Auth,
	})
}

// validateEndpoint は、endpoint がプッシュサーバーの URL として使えるかを確かめる。
// サーバーはこの URL に通知を送るので、https 以外や、IP アドレス・localhost のような
// 内部のサーバーを指す URL は受け付けない（プッシュサーバーはどれもドメイン名で、https）。
func validateEndpoint(endpoint string) error {
	invalid := domain.InvalidInput("endpoint は https のプッシュサーバーの URL にしてください")
	if endpoint == "" || len(endpoint) > maxEndpointLength {
		return invalid
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return invalid
	}
	host := u.Hostname()
	if host == "" || net.ParseIP(host) != nil || !strings.Contains(host, ".") ||
		host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return invalid
	}
	return nil
}

// validPushKey は、鍵が base64url（末尾の = はあってもなくてもよい）で、決まったバイト数かを確かめる。
func validPushKey(key string, size int) bool {
	if key == "" || len(key) > maxPushKeyLength {
		return false
	}
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(key, "="))
	return err == nil && len(b) == size
}
