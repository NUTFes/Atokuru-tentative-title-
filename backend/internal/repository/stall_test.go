package repository

import "testing"

func TestListWithWaitingCount_UsesIndex(t *testing.T) {
	gdb := newTestDB(t)
	assertUsesIndex(t, gdb, "idx_tickets_stall_status_order", listWithWaitingCountSQL, "waiting")
}
