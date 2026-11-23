package repository

import (
	"context"
	"slices"
	"strconv"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

func TestMaxReceivedOrderNumber(t *testing.T) {
	gdb := newTestDB(t)
	ctx := context.Background()
	tickets := NewTicketRepository(gdb)

	s := domain.Stall{Name: "屋台", StallNumber: "A-01", BusinessStatus: domain.BusinessStatusOpen}
	other := domain.Stall{Name: "屋台2", StallNumber: "A-02", BusinessStatus: domain.BusinessStatusOpen}
	if err := gdb.Create(&s).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&other).Error; err != nil {
		t.Fatal(err)
	}

	if got, err := tickets.MaxReceivedOrderNumber(ctx, s.ID); err != nil || got != 0 {
		t.Fatalf("受け取り済みがないとき: got %d, err %v", got, err)
	}

	for i, st := range []domain.TicketStatus{"received", "received", "waiting", "received", "ready"} {
		tk := domain.Ticket{PublicToken: "s" + string(rune('a'+i)), StallID: s.ID, OrderNumber: i + 1, Status: st}
		if err := gdb.Create(&tk).Error; err != nil {
			t.Fatal(err)
		}
	}
	// 別の屋台の受け取り済みは数えない
	if err := gdb.Create(&domain.Ticket{PublicToken: "o1", StallID: other.ID, OrderNumber: 100, Status: "received"}).Error; err != nil {
		t.Fatal(err)
	}

	if got, err := tickets.MaxReceivedOrderNumber(ctx, s.ID); err != nil || got != 4 {
		t.Errorf("got %d, err %v, want 4", got, err)
	}

	assertUsesIndex(t, gdb, "idx_tickets_stall_status_order",
		"SELECT COALESCE(MAX(order_number), 0) FROM tickets WHERE stall_id = ? AND status = ?", s.ID, "received")
}

func TestListByStall_UsesIndex(t *testing.T) {
	gdb := newTestDB(t)
	tickets := NewTicketRepository(gdb)
	stmt := tickets.listByStallQuery(context.Background(), 1, []domain.TicketStatus{"waiting", "ready"}, 100, false).
		Session(&gorm.Session{DryRun: true}).
		Find(&[]domain.Ticket{}).Statement
	assertUsesIndex(t, gdb, "idx_tickets_stall_status_order", stmt.SQL.String(), stmt.Vars...)
}

func TestClaimNearNotification(t *testing.T) {
	gdb := newTestDB(t)
	tickets := NewTicketRepository(gdb)
	ctx := context.Background()
	stall := domain.Stall{Name: "屋台", StallNumber: "A-01", BusinessStatus: domain.BusinessStatusOpen}
	if err := gdb.Create(&stall).Error; err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 6; i++ {
		if _, err := tickets.Create(ctx, stall.ID, "token"+strconv.Itoa(i), nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	// 4番は調理完了にしておく（調理中でないので選ばれない）。
	gdb.Model(&domain.Ticket{}).Where("order_number = 4").Update("status", domain.TicketStatusReady)

	now := time.Now()
	got, err := tickets.ClaimNearNotification(ctx, stall.ID, 3, 5, now)
	if err != nil {
		t.Fatal(err)
	}
	var nums []int
	for _, tk := range got {
		nums = append(nums, tk.OrderNumber)
	}
	if !slices.Equal(nums, []int{3, 5}) {
		t.Errorf("選ばれた番号 = %v, want [3 5]", nums)
	}

	// 2回目は、もう送ったことになっているので選ばれない。
	got, err = tickets.ClaimNearNotification(ctx, stall.ID, 1, 6, now)
	if err != nil {
		t.Fatal(err)
	}
	nums = nums[:0]
	for _, tk := range got {
		nums = append(nums, tk.OrderNumber)
	}
	if !slices.Equal(nums, []int{1, 2, 6}) {
		t.Errorf("2回目に選ばれた番号 = %v, want [1 2 6]", nums)
	}

	// from > to なら何も選ばない。
	if got, err := tickets.ClaimNearNotification(ctx, stall.ID, 5, 4, now); err != nil || len(got) != 0 {
		t.Errorf("from > to: got %v, err %v", got, err)
	}
}

func TestClaimNearNotification_UsesIndex(t *testing.T) {
	gdb := newTestDB(t)
	tickets := NewTicketRepository(gdb)
	stmt := tickets.nearCandidatesQuery(context.Background(), 1, 5, 8).
		Session(&gorm.Session{DryRun: true}).
		Find(&[]domain.Ticket{}).Statement
	assertUsesIndex(t, gdb, "idx_tickets_stall_status_order", stmt.SQL.String(), stmt.Vars...)
}
