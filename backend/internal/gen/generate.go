// Package gen は openapi/openapi.yaml から oapi-codegen で生成したコードを置く。
// api.gen.go は手で編集せず、仕様書を直してから backend/ で `go generate ./...` を実行する。
package gen

//go:generate go tool oapi-codegen -config ../../oapi-codegen.yaml ../../../openapi/openapi.yaml
