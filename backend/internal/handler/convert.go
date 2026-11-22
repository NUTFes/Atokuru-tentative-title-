package handler

import (
	"math"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/usecase"
)

// domain の entity を、仕様書の形（gen の型）に詰め替える。

func toStall(s *domain.Stall) gen.Stall {
	return gen.Stall{
		Id:             int64(s.ID),
		Name:           s.Name,
		StallNumber:    s.StallNumber,
		HpLink:         s.HPLink,
		BusinessStatus: gen.BusinessStatus(s.BusinessStatus),
	}
}

func toTicket(t *domain.Ticket) gen.Ticket {
	return gen.Ticket{
		Id:          int64(t.ID),
		PublicToken: t.PublicToken,
		StallId:     int64(t.StallID),
		OrderNumber: t.OrderNumber,
		Status:      gen.TicketStatus(t.Status),
		Comment:     t.Comment,
		CalledAt:    t.CalledAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toTicketDetail(d *usecase.TicketDetail) gen.TicketDetail {
	return gen.TicketDetail{
		PublicToken: d.Ticket.PublicToken,
		StallId:     int64(d.Ticket.StallID),
		StallName:   d.Stall.Name,
		StallNumber: d.Stall.StallNumber,
		OrderNumber: d.Ticket.OrderNumber,
		Status:      gen.TicketStatus(d.Ticket.Status),
		Comment:     d.Ticket.Comment,
		CalledAt:    d.Ticket.CalledAt,
		CreatedAt:   d.Ticket.CreatedAt,
		PeopleAhead: d.PeopleAhead,
	}
}

// toID はパスの ID を uint にする。0 以下や大きすぎる値は、存在しない ID として 0 を返す（0 の ID はない）。
func toID(id int64) uint {
	if id <= 0 || uint64(id) > math.MaxUint32 {
		return 0
	}
	return uint(id)
}
