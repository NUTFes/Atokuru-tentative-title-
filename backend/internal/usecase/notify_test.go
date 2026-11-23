package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"testing"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/infra/db"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/repository"
)

func TestNearRange(t *testing.T) {
	tests := []struct {
		before, after int
		lo, hi        int
	}{
		{0, 1, 5, 5},    // 1番を受け取り → 5番が前に3人になる（2〜4番は発券時から3人以下）
		{1, 2, 6, 6},    // 1つ進む → 1人ずつ
		{2, 5, 7, 9},    // 3つ進む → 7〜9番
		{2, 20, 21, 24}, // 大きく進む → 新しい最大番号のすぐ後ろの4人だけ
		{5, 5, 1, 0},    // 進まない → なし
		{5, 3, 1, 0},    // 戻る（最大より前の番号を受け取った）→ なし
	}
	for _, tt := range tests {
		lo, hi := nearRange(tt.before, tt.after)
		if tt.lo > tt.hi {
			if lo <= hi {
				t.Errorf("nearRange(%d, %d) = %d〜%d, want なし", tt.before, tt.after, lo, hi)
			}
			continue
		}
		if lo != tt.lo || hi != tt.hi {
			t.Errorf("nearRange(%d, %d) = %d〜%d, want %d〜%d", tt.before, tt.after, lo, hi, tt.lo, tt.hi)
		}
	}
}

// fakeSender は送った通知を覚えておく。gone に入っている endpoint には ErrPushSubscriptionGone を返す。
type fakeSender struct {
	mu   sync.Mutex
	sent []sentPush
	gone map[string]bool
}

type sentPush struct {
	endpoint string
	msg      pushMessage
}

func (f *fakeSender) Send(_ context.Context, sub domain.PushSubscription, payload []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	var m pushMessage
	if err := json.Unmarshal(payload, &m); err != nil {
		return err
	}
	f.sent = append(f.sent, sentPush{endpoint: sub.Endpoint, msg: m})
	if f.gone[sub.Endpoint] {
		return domain.ErrPushSubscriptionGone
	}
	return nil
}

// take は、これまでに送った通知を返して空にする。
func (f *fakeSender) take() []sentPush {
	f.mu.Lock()
	defer f.mu.Unlock()
	s := f.sent
	f.sent = nil
	return s
}

type notifyFixture struct {
	gdb     *gorm.DB
	usecase *TicketUsecase
	sender  *fakeSender
	stall   *domain.Stall
	tickets []*domain.Ticket // tickets[i] は i+1 番
}

// newNotifyFixture は、営業中の屋台に n 枚の整理券を発行し、全部に購読を1件ずつ付ける。
// 通知はすぐに（同期で）送るようにする。
func newNotifyFixture(t *testing.T, n int) *notifyFixture {
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
	sender := &fakeSender{gone: map[string]bool{}}
	u := NewTicketUsecase(stallRepo, ticketRepo, pushRepo, sender)
	u.runAsync = func(f func()) { f() }

	stall := &domain.Stall{Name: "焼きそば研究会", StallNumber: "A-01", BusinessStatus: domain.BusinessStatusOpen}
	if err := gdb.Create(stall).Error; err != nil {
		t.Fatal(err)
	}
	fx := &notifyFixture{gdb: gdb, usecase: u, sender: sender, stall: stall}
	for i := 0; i < n; i++ {
		tk, err := u.Create(t.Context(), stall.ID, nil)
		if err != nil {
			t.Fatal(err)
		}
		fx.tickets = append(fx.tickets, tk)
		fx.subscribe(t, tk, "https://push.example.com/"+tk.PublicToken)
	}
	return fx
}

func (fx *notifyFixture) subscribe(t *testing.T, tk *domain.Ticket, endpoint string) {
	t.Helper()
	sub := &domain.PushSubscription{TicketID: tk.ID, Endpoint: endpoint, P256dh: "p", Auth: "a"}
	if err := fx.gdb.Create(sub).Error; err != nil {
		t.Fatal(err)
	}
}

// set は order 番の整理券の status を変える。
func (fx *notifyFixture) set(t *testing.T, order int, status domain.TicketStatus) {
	t.Helper()
	if _, err := fx.usecase.UpdateStatus(t.Context(), fx.tickets[order-1].ID, status); err != nil {
		t.Fatal(err)
	}
}

// sentOrders は、送った通知の種類と番号を "near:5" の形で並べる。
func sentOrders(sent []sentPush) []string {
	var s []string
	for _, p := range sent {
		s = append(s, p.msg.Type+":"+strconv.Itoa(p.msg.OrderNumber))
	}
	return s
}

func TestNotify_Ready(t *testing.T) {
	fx := newNotifyFixture(t, 3)
	tk := fx.tickets[1]
	fx.subscribe(t, tk, "https://push.example.com/second-device")

	fx.set(t, 2, domain.TicketStatusReady)
	sent := fx.sender.take()
	if len(sent) != 2 {
		t.Fatalf("送った数 = %d, want 2（2番の端末2台）: %v", len(sent), sentOrders(sent))
	}
	want := pushMessage{
		Type: "ready", PublicToken: tk.PublicToken, StallName: "焼きそば研究会", StallNumber: "A-01",
		OrderNumber: 2, PeopleAhead: 0, URL: "/?token=" + tk.PublicToken,
	}
	for _, p := range sent {
		if p.msg != want {
			t.Errorf("通知の中身 = %+v, want %+v", p.msg, want)
		}
	}

	// 同じ status をもう一度指定しても送らない。
	fx.set(t, 2, domain.TicketStatusReady)
	if sent := fx.sender.take(); len(sent) != 0 {
		t.Errorf("同じ status で送っています: %v", sentOrders(sent))
	}

	// 受け取り済みを押し間違いで調理完了に戻したときは、もう呼び出し済みなので送らない。
	fx.set(t, 2, domain.TicketStatusReceived)
	fx.sender.take()
	fx.set(t, 2, domain.TicketStatusReady)
	if sent := fx.sender.take(); len(sent) != 0 {
		t.Errorf("received から ready に戻したときに送っています: %v", sentOrders(sent))
	}
}

func TestNotify_Near(t *testing.T) {
	fx := newNotifyFixture(t, 12)

	steps := []struct {
		name   string
		order  int
		status domain.TicketStatus
		want   []string
	}{
		// 2〜4番は発券した時点で前に3人以下なので送らない。5番が初めて前に3人になる。
		{"1番を受け取り", 1, domain.TicketStatusReceived, []string{"near:5"}},
		{"2番を受け取り", 2, domain.TicketStatusReceived, []string{"near:6"}},
		// 6番を飛ばして5番を受け取ると、7〜9番が一度に3人以下になる。8番は調理完了なので除く。
		{"8番を調理完了", 8, domain.TicketStatusReady, []string{"ready:8"}},
		{"5番を受け取り", 5, domain.TicketStatusReceived, []string{"near:7", "near:9"}},
		// 最大番号より前の番号を受け取っても、番号は進まないので送らない。
		{"3番を受け取り", 3, domain.TicketStatusReceived, nil},
		// 押し間違いで5番を戻し、もう一度受け取っても、7・9番には2回目を送らない。
		{"5番を調理中に戻す", 5, domain.TicketStatusWaiting, nil},
		{"5番をもう一度受け取り", 5, domain.TicketStatusReceived, nil},
		{"6番を受け取り", 6, domain.TicketStatusReceived, []string{"near:10"}},
	}
	for _, s := range steps {
		fx.set(t, s.order, s.status)
		if got := sentOrders(fx.sender.take()); !slices.Equal(got, s.want) {
			t.Errorf("%s: 送った通知 = %v, want %v", s.name, got, s.want)
		}
	}

	// 送った「あと3人」の前に何人は、その時点の値になっている。
	fx.set(t, 7, domain.TicketStatusReceived) // 最大 6 → 7、11番が前に3人
	sent := fx.sender.take()
	if len(sent) != 1 || sent[0].msg.OrderNumber != 11 || sent[0].msg.PeopleAhead != 3 {
		t.Errorf("7番を受け取り: %+v, want 11番・前に3人", sent)
	}
}

func TestNotify_RemovesGoneSubscription(t *testing.T) {
	fx := newNotifyFixture(t, 1)
	gone := "https://push.example.com/" + fx.tickets[0].PublicToken
	fx.sender.gone[gone] = true

	fx.set(t, 1, domain.TicketStatusReady)
	var subs []domain.PushSubscription
	fx.gdb.Find(&subs)
	if len(subs) != 0 {
		t.Errorf("使えなくなった購読が残っています: %+v", subs)
	}
}

// 通知の送信に失敗しても、status の変更は成功する。
func TestNotify_SendErrorDoesNotFailUpdate(t *testing.T) {
	fx := newNotifyFixture(t, 1)
	fx.usecase.sender = senderFunc(func(context.Context, domain.PushSubscription, []byte) error {
		return errors.New("プッシュサーバーが落ちている")
	})
	got, err := fx.usecase.UpdateStatus(t.Context(), fx.tickets[0].ID, domain.TicketStatusReady)
	if err != nil || got.Status != domain.TicketStatusReady {
		t.Errorf("got %+v, err %v", got, err)
	}
}

// 鍵がない（sender が nil の）ときは、通知を送らずに status だけ変える。
func TestNotify_Disabled(t *testing.T) {
	fx := newNotifyFixture(t, 5)
	fx.usecase.sender = nil
	fx.set(t, 1, domain.TicketStatusReceived)
	fx.set(t, 2, domain.TicketStatusReady)
	if sent := fx.sender.take(); len(sent) != 0 {
		t.Errorf("鍵がないのに送っています: %v", sentOrders(sent))
	}
}

type senderFunc func(context.Context, domain.PushSubscription, []byte) error

func (f senderFunc) Send(ctx context.Context, sub domain.PushSubscription, payload []byte) error {
	return f(ctx, sub, payload)
}
