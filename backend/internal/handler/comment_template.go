package handler

import (
	"context"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

// ListCommentTemplates は GET /stalls/{stallId}/comment-templates。
func (h *Handler) ListCommentTemplates(ctx context.Context, req gen.ListCommentTemplatesRequestObject) (gen.ListCommentTemplatesResponseObject, error) {
	ts, err := h.templates.List(ctx, toID(req.StallId))
	if err != nil {
		return nil, err
	}
	res := make(gen.ListCommentTemplates200JSONResponse, 0, len(ts))
	for i := range ts {
		res = append(res, toCommentTemplate(&ts[i]))
	}
	return res, nil
}

// CreateCommentTemplate は POST /stalls/{stallId}/comment-templates。
func (h *Handler) CreateCommentTemplate(ctx context.Context, req gen.CreateCommentTemplateRequestObject) (gen.CreateCommentTemplateResponseObject, error) {
	if req.Body == nil {
		return nil, domain.InvalidInput("定型文を入力してください")
	}
	t, err := h.templates.Create(ctx, toID(req.StallId), req.Body.Comment)
	if err != nil {
		return nil, err
	}
	return gen.CreateCommentTemplate201JSONResponse(toCommentTemplate(t)), nil
}

// DeleteCommentTemplate は DELETE /comment-templates/{templateId}。
func (h *Handler) DeleteCommentTemplate(ctx context.Context, req gen.DeleteCommentTemplateRequestObject) (gen.DeleteCommentTemplateResponseObject, error) {
	if err := h.templates.Delete(ctx, toID(req.TemplateId)); err != nil {
		return nil, err
	}
	return gen.DeleteCommentTemplate204Response{}, nil
}
