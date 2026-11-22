package handler

import (
	"context"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

// ListStalls は GET /stalls（マップ用の一覧）。
func (h *Handler) ListStalls(ctx context.Context, _ gen.ListStallsRequestObject) (gen.ListStallsResponseObject, error) {
	return nil, errNotImplemented
}

// GetStall は GET /stalls/{stallId}。
func (h *Handler) GetStall(ctx context.Context, req gen.GetStallRequestObject) (gen.GetStallResponseObject, error) {
	return nil, errNotImplemented
}

// UpdateBusinessStatus は PATCH /stalls/{stallId}/business-status。
func (h *Handler) UpdateBusinessStatus(ctx context.Context, req gen.UpdateBusinessStatusRequestObject) (gen.UpdateBusinessStatusResponseObject, error) {
	return nil, errNotImplemented
}
