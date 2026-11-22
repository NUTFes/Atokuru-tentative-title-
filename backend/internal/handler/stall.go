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
	s, err := h.stalls.Get(ctx, toID(req.StallId))
	if err != nil {
		return nil, err
	}
	return gen.GetStall200JSONResponse(toStall(s)), nil
}

// UpdateBusinessStatus は PATCH /stalls/{stallId}/business-status。
func (h *Handler) UpdateBusinessStatus(ctx context.Context, req gen.UpdateBusinessStatusRequestObject) (gen.UpdateBusinessStatusResponseObject, error) {
	return nil, errNotImplemented
}
