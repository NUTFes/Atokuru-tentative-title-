package usecase

import (
	"testing"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

func TestPeopleAhead(t *testing.T) {
	tests := []struct {
		name        string
		orderNumber int
		status      domain.TicketStatus
		maxReceived int
		want        int
	}{
		{"まだ誰も受け取っていない・1番", 1, domain.TicketStatusWaiting, 0, 0},
		{"まだ誰も受け取っていない・5番", 5, domain.TicketStatusWaiting, 0, 4},
		{"3番まで受け取り済み・5番", 5, domain.TicketStatusWaiting, 3, 1},
		{"4番まで受け取り済み・5番", 5, domain.TicketStatusWaiting, 4, 0},
		{"後ろの番号が先に受け取った", 5, domain.TicketStatusWaiting, 8, 0},
		{"調理完了なら 0", 5, domain.TicketStatusReady, 0, 0},
		{"受け取り済みなら 0", 5, domain.TicketStatusReceived, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := peopleAhead(&domain.Ticket{OrderNumber: tt.orderNumber, Status: tt.status}, tt.maxReceived)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
