package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

// GetVapidPublicKey は GET /push/vapid-public-key。#60 で実装する。
func (h *Handler) GetVapidPublicKey(ctx context.Context, req gen.GetVapidPublicKeyRequestObject) (gen.GetVapidPublicKeyResponseObject, error) {
	return nil, echo.NewHTTPError(http.StatusNotImplemented)
}

// CreatePushSubscription は POST /tickets/{publicToken}/push-subscriptions。#60 で実装する。
func (h *Handler) CreatePushSubscription(ctx context.Context, req gen.CreatePushSubscriptionRequestObject) (gen.CreatePushSubscriptionResponseObject, error) {
	return nil, echo.NewHTTPError(http.StatusNotImplemented)
}
