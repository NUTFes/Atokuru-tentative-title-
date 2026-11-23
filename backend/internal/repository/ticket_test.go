package repository

import (
	"context"
	"testing"

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
