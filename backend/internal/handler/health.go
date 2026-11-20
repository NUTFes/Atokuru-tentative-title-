package handler

import (
	"context"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

// GetHealth は GET /health。死活監視に使う。
func (h *Handler) GetHealth(ctx context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	return gen.GetHealth200JSONResponse{Status: gen.Ok}, nil
}
