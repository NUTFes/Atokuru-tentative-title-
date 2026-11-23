package handler

import (
	"context"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/usecase"
)

// ListStallTickets は GET /stalls/{stallId}/tickets。
func (h *Handler) ListStallTickets(ctx context.Context, req gen.ListStallTicketsRequestObject) (gen.ListStallTicketsResponseObject, error) {
	in := usecase.ListInput{Limit: req.Params.Limit}
	for _, s := range req.Params.Status {
		in.Statuses = append(in.Statuses, domain.TicketStatus(s))
	}
	if req.Params.Order != nil {
		in.Order = string(*req.Params.Order)
	}

	tickets, err := h.tickets.ListByStall(ctx, toID(req.StallId), in)
	if err != nil {
		return nil, err
	}
	res := make(gen.ListStallTickets200JSONResponse, 0, len(tickets))
	for i := range tickets {
		res = append(res, toTicket(&tickets[i]))
	}
	return res, nil
}

// CreateTicket は POST /stalls/{stallId}/tickets。
func (h *Handler) CreateTicket(ctx context.Context, req gen.CreateTicketRequestObject) (gen.CreateTicketResponseObject, error) {
	var comment *string
	if req.Body != nil {
		comment = req.Body.Comment
	}
	t, err := h.tickets.Create(ctx, toID(req.StallId), comment)
	if err != nil {
		return nil, err
	}
	return gen.CreateTicket201JSONResponse(toTicket(t)), nil
}

// GetTicket は GET /tickets/{publicToken}。
func (h *Handler) GetTicket(ctx context.Context, req gen.GetTicketRequestObject) (gen.GetTicketResponseObject, error) {
	d, err := h.tickets.GetByPublicToken(ctx, req.PublicToken)
	if err != nil {
		return nil, err
	}
	return gen.GetTicket200JSONResponse(toTicketDetail(d)), nil
}

// UpdateTicketStatus は PATCH /tickets/{ticketId}/status。
func (h *Handler) UpdateTicketStatus(ctx context.Context, req gen.UpdateTicketStatusRequestObject) (gen.UpdateTicketStatusResponseObject, error) {
	if req.Body == nil {
		return nil, domain.InvalidInput("status を指定してください")
	}
	t, err := h.tickets.UpdateStatus(ctx, toID(req.TicketId), domain.TicketStatus(req.Body.Status))
	if err != nil {
		return nil, err
	}
	return gen.UpdateTicketStatus200JSONResponse(toTicket(t)), nil
}
