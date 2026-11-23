package handler

import (
	"context"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/usecase"
)

// GetVapidPublicKey は GET /push/vapid-public-key。
func (h *Handler) GetVapidPublicKey(ctx context.Context, req gen.GetVapidPublicKeyRequestObject) (gen.GetVapidPublicKeyResponseObject, error) {
	key, err := h.push.VapidPublicKey()
	if err != nil {
		return nil, err
	}
	return gen.GetVapidPublicKey200JSONResponse{PublicKey: key}, nil
}

// CreatePushSubscription は POST /tickets/{publicToken}/push-subscriptions。
func (h *Handler) CreatePushSubscription(ctx context.Context, req gen.CreatePushSubscriptionRequestObject) (gen.CreatePushSubscriptionResponseObject, error) {
	if req.Body == nil {
		return nil, domain.InvalidInput("購読の内容を指定してください")
	}
	err := h.push.Subscribe(ctx, req.PublicToken, usecase.SubscribeInput{
		Endpoint: req.Body.Endpoint,
		P256dh:   req.Body.Keys.P256dh,
		Auth:     req.Body.Keys.Auth,
	})
	if err != nil {
		return nil, err
	}
	return gen.CreatePushSubscription204Response{}, nil
}
