package handler

import (
	"context"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

// ListStallTickets は GET /stalls/{stallId}/tickets。
func (h *Handler) ListStallTickets(ctx context.Context, req gen.ListStallTicketsRequestObject) (gen.ListStallTicketsResponseObject, error) {
	return nil, errNotImplemented
}

// CreateTicket は POST /stalls/{stallId}/tickets。
func (h *Handler) CreateTicket(ctx context.Context, req gen.CreateTicketRequestObject) (gen.CreateTicketResponseObject, error) {
	return nil, errNotImplemented
}

// GetTicket は GET /tickets/{publicToken}。
func (h *Handler) GetTicket(ctx context.Context, req gen.GetTicketRequestObject) (gen.GetTicketResponseObject, error) {
	return nil, errNotImplemented
}

// UpdateTicketStatus は PATCH /tickets/{ticketId}/status。
func (h *Handler) UpdateTicketStatus(ctx context.Context, req gen.UpdateTicketStatusRequestObject) (gen.UpdateTicketStatusResponseObject, error) {
	return nil, errNotImplemented
}
