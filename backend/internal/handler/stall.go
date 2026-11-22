package handler

import (
	"context"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
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
	if req.Body == nil {
		return nil, domain.InvalidInput("business_status を指定してください")
	}
	s, err := h.stalls.UpdateBusinessStatus(ctx, toID(req.StallId), domain.BusinessStatus(req.Body.BusinessStatus))
	if err != nil {
		return nil, err
	}
	return gen.UpdateBusinessStatus200JSONResponse(toStall(s)), nil
}
