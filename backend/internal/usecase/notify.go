package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// PushSender は通知（Web Push）を購読1件に送る。実装は infra/notify にあり、main.go で渡す。
// プッシュサーバーが「もう使えない購読」と返したときは domain.ErrPushSubscriptionGone を返す。
type PushSender interface {
	Send(ctx context.Context, sub domain.PushSubscription, payload []byte) error
}

// 通知の種類（仕様書の info.description に書いた push の data の type）。
const (
	pushTypeReady = "ready" // 調理完了
	pushTypeNear  = "near"  // あと3人
)

// nearThreshold は「あと3人」の通知を送る、前に何人の上限。
const nearThreshold = 3

// notifyTimeout は、status の変更1回ぶんの通知を送り終えるまでの上限の時間。
const notifyTimeout = 30 * time.Second

// pushMessage は通知の中身（push の data）。文言は FE の sw.js が type を見て組み立てる。
type pushMessage struct {
	Type        string `json:"type"`
	PublicToken string `json:"public_token"`
	StallName   string `json:"stall_name"`
	StallNumber string `json:"stall_number"`
	OrderNumber int    `json:"order_number"`
	PeopleAhead int    `json:"people_ahead"`
	URL         string `json:"url"`
}

// ticketURL は、通知を押したときに開く整理券の画面の URL（QR コードと同じ形）。
func ticketURL(publicToken string) string {
	return "/?token=" + publicToken
}

// nearRange は、受け取り済みの最大番号が before から after に進んだときに、
// 「あと3人」を新しく送る整理券の order_number の範囲を返す（lo > hi なら送る相手はいない）。
//
// 前に何人 = 番号 − after − 1 が nearThreshold 以下になるのは、番号が after+1 〜 after+nearThreshold+1 の整理券。
// そのうち、進む前から3人以下だった（番号が before+nearThreshold+1 以下の）整理券は除く。
// 発券した時点ですでに3人以下の人は画面で分かるので送らず、送るのは番号が進んで3人以下になった人だけにする。
func nearRange(before, after int) (lo, hi int) {
	if after <= before {
		return 1, 0
	}
	return max(after+1, before+nearThreshold+2), after + nearThreshold + 1
}

// notifyStatusChange は、整理券の status が from から t.Status に変わったあとの通知を送る。
// maxReceivedBefore は、変える前の、その屋台の受け取り済みの最大番号。
func (u *TicketUsecase) notifyStatusChange(ctx context.Context, t *domain.Ticket, from domain.TicketStatus, maxReceivedBefore int) {
	switch t.Status {
	case domain.TicketStatusReady:
		// 調理中から調理完了にしたときだけ送る。受け取り済みを押し間違いで戻したときは、もう呼び出し済みなので送らない。
		if from != domain.TicketStatusWaiting {
			return
		}
		u.sendToTickets(ctx, pushTypeReady, t.StallID, []domain.Ticket{*t}, 0)

	case domain.TicketStatusReceived:
		maxReceived := max(maxReceivedBefore, t.OrderNumber)
		lo, hi := nearRange(maxReceivedBefore, maxReceived)
		near, err := u.tickets.ClaimNearNotification(ctx, t.StallID, lo, hi, time.Now())
		if err != nil {
			log.Printf("あと3人の通知の相手を決められません（屋台 %d）: %v", t.StallID, err)
		}
		u.sendToTickets(ctx, pushTypeNear, t.StallID, near, maxReceived)
	}
}

// sendToTickets は、整理券それぞれの購読すべてに通知を送る。
// 失敗してもログに出すだけにする（status の変更はもう終わっているので、取り消さない）。
func (u *TicketUsecase) sendToTickets(ctx context.Context, typ string, stallID uint, tickets []domain.Ticket, maxReceived int) {
	if len(tickets) == 0 {
		return
	}
	ids := make([]uint, len(tickets))
	byID := make(map[uint]*domain.Ticket, len(tickets))
	for i := range tickets {
		ids[i] = tickets[i].ID
		byID[tickets[i].ID] = &tickets[i]
	}
	subs, err := u.subscriptions.ListByTicketIDs(ctx, ids)
	if err != nil {
		log.Printf("通知の購読を読めません: %v", err)
		return
	}
	if len(subs) == 0 {
		return
	}
	stall, err := u.stalls.FindByID(ctx, stallID)
	if err != nil {
		log.Printf("通知に載せる屋台を読めません（屋台 %d）: %v", stallID, err)
		return
	}

	for _, sub := range subs {
		t := byID[sub.TicketID]
		// 前に何人は GET /tickets/{publicToken} と同じ計算にする（調理完了なら 0）。
		payload, err := json.Marshal(pushMessage{
			Type:        typ,
			PublicToken: t.PublicToken,
			StallName:   stall.Name,
			StallNumber: stall.StallNumber,
			OrderNumber: t.OrderNumber,
			PeopleAhead: peopleAhead(t, maxReceived),
			URL:         ticketURL(t.PublicToken),
		})
		if err != nil {
			log.Printf("通知の中身を作れません: %v", err)
			continue
		}

		err = u.sender.Send(ctx, sub, payload)
		switch {
		case errors.Is(err, domain.ErrPushSubscriptionGone):
			// 端末で通知をやめた、アプリを消したなど。以降は送らないので消す。
			if err := u.subscriptions.Delete(ctx, sub.ID); err != nil {
				log.Printf("使えなくなった購読を消せません（購読 %d）: %v", sub.ID, err)
			}
		case err != nil:
			log.Printf("通知を送れません（整理券 %d、購読 %d）: %v", sub.TicketID, sub.ID, err)
		}
	}
}
