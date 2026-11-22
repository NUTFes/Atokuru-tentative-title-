package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// isAPIPath は API の URL（/api の下）かを返す。
func isAPIPath(path string) bool {
	return path == "/api" || strings.HasPrefix(path, "/api/")
}

// ServeFrontend は、/api 以外の URL で FE（Next.js の静的エクスポート）のファイルを返すようにする。
// dir は `npm run build` で出力された out/ のディレクトリ。
//
// next.config の trailingSlash: true で /map は out/map/index.html になるので、
// ディレクトリの URL では index.html を返す。ないページは out/404.html を 404 で返す。
func ServeFrontend(e *echo.Echo, dir string) {
	e.Use(frontendCacheControl)
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:    dir,
		Skipper: func(c echo.Context) bool { return isAPIPath(c.Request().URL.Path) },
	}))

	apiErrorHandler := e.HTTPErrorHandler
	notFoundPage := filepath.Join(dir, "404.html")
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if !isAPIPath(c.Request().URL.Path) && isNotFound(err) {
			if page, readErr := os.ReadFile(notFoundPage); readErr == nil {
				_ = c.HTMLBlob(http.StatusNotFound, page)
				return
			}
		}
		apiErrorHandler(err, c)
	}
}

// frontendCacheControl は、ファイル名にハッシュが付いた /_next/static/ は長くキャッシュさせ、
// それ以外（HTML など）は毎回サーバーに確かめさせる。HTML が古いままだと更新が届かないため。
func frontendCacheControl(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
		switch {
		case isAPIPath(path):
			c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
		case strings.HasPrefix(path, "/_next/static/"):
			c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=31536000, immutable")
		default:
			c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		}
		return next(c)
	}
}

// AllowCORS は、origins からの API の呼び出しを許可する。
// 本番は FE と API が同じオリジンなので使わない。ローカルで `next dev`（3000番）から
// API（8080番）を呼ぶときだけ、環境変数 CORS_ALLOW_ORIGINS で指定する。
func AllowCORS(e *echo.Echo, origins []string) {
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType},
	}))
}

func isNotFound(err error) bool {
	status, _ := toErrorResponse(err)
	return status == http.StatusNotFound
}
