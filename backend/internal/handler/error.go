package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

// HTTPErrorHandler は、どの API で起きたエラーも仕様書の Error（code と message）の形で返す。
// handler が返したエラーも、Echo がルーティングや引数の読み取りで出したエラーも、すべてここを通る。
func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status, body := toErrorResponse(err)
	if status >= http.StatusInternalServerError {
		c.Logger().Error(err)
	}
	if c.Request().Method == http.MethodHead {
		_ = c.NoContent(status)
		return
	}
	_ = c.JSON(status, body)
}

func toErrorResponse(err error) (int, gen.Error) {
	var de *domain.Error
	if errors.As(err, &de) {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return http.StatusNotFound, gen.Error{Code: "not_found", Message: de.Message}
		case errors.Is(err, domain.ErrInvalidInput):
			return http.StatusBadRequest, gen.Error{Code: "bad_request", Message: de.Message}
		case errors.Is(err, domain.ErrStallNotOpen):
			return http.StatusConflict, gen.Error{Code: "stall_not_open", Message: de.Message}
		case errors.Is(err, domain.ErrPushUnavailable):
			return http.StatusServiceUnavailable, gen.Error{Code: "push_unavailable", Message: de.Message}
		}
	}

	var he *echo.HTTPError
	if errors.As(err, &he) {
		switch {
		case he.Code == http.StatusNotFound:
			return he.Code, gen.Error{Code: "not_found", Message: "見つかりません"}
		case he.Code == http.StatusMethodNotAllowed:
			return he.Code, gen.Error{Code: "method_not_allowed", Message: "このメソッドは使えません"}
		case he.Code >= 400 && he.Code < 500:
			// パスやクエリ、JSON の形が仕様書と合わないときなど
			return he.Code, gen.Error{Code: "bad_request", Message: fmt.Sprint(he.Message)}
		}
	}

	return http.StatusInternalServerError, gen.Error{Code: "internal_error", Message: "サーバーでエラーが起きました"}
}
