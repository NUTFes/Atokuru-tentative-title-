package usecase

import (
	"testing"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

func TestCongestion(t *testing.T) {
	tests := []struct {
		status  domain.BusinessStatus
		waiting int
		want    domain.Congestion
	}{
		{domain.BusinessStatusOpen, 0, domain.CongestionGreen},
		{domain.BusinessStatusOpen, congestionYellowFrom - 1, domain.CongestionGreen},
		{domain.BusinessStatusOpen, congestionYellowFrom, domain.CongestionYellow},
		{domain.BusinessStatusOpen, congestionRedFrom - 1, domain.CongestionYellow},
		{domain.BusinessStatusOpen, congestionRedFrom, domain.CongestionRed},
		{domain.BusinessStatusPreparing, 0, domain.CongestionGray},
		{domain.BusinessStatusSoldOut, 100, domain.CongestionGray},
		{domain.BusinessStatusClosed, 100, domain.CongestionGray},
	}
	for _, tt := range tests {
		if got := congestion(tt.status, tt.waiting); got != tt.want {
			t.Errorf("congestion(%s, %d) = %s, want %s", tt.status, tt.waiting, got, tt.want)
		}
	}
}
