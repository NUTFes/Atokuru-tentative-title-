// Package translate は、備考の定型文を OpenAI の API で訳す。
package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// defaultBaseURL は OpenAI の API の URL。テストでは手元のサーバーに差し替える。
const defaultBaseURL = "https://api.openai.com/v1"

// OpenAI は OpenAI の Chat Completions API で訳す。
// API キーはサーバーの環境変数にだけ置き、FE には出さない。
type OpenAI struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewOpenAI は翻訳の準備をする。model は OPENAI_MODEL に書いたモデル名。
func NewOpenAI(apiKey, model string) *OpenAI {
	return &OpenAI{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

// systemPrompt は、何を訳すのかと、返す形を伝える指示。
// 文脈と訳語を決めておくと、訳がぶれにくい（会話ログの「精度を上げるコツ」）。
const systemPrompt = `あなたは大学祭の屋台の整理券アプリの翻訳者です。
屋台が整理券の備考に使う定型文（メニュー名や注文の補足。例：「焼きそば 大盛り」「ソース抜き」）を、来場者（留学生や海外からの来場者）向けに訳します。

- 指定された言語コードをキーにした JSON オブジェクトだけを返してください。例：{"en": "Yakisoba, large"}
- スマホの小さな欄に出すので、短く訳してください
- 日本の料理名で一般的な英語名がないもの（焼きそば、たこ焼きなど）はローマ字にしてください
- 説明や注釈は付けないでください`

// Translate は text を langs の言語に訳す。
func (o *OpenAI) Translate(ctx context.Context, text string, langs []string) (domain.Translations, error) {
	user, err := json.Marshal(map[string]any{"languages": langs, "text": text})
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(map[string]any{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": string(user)},
		},
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI の API を呼べません: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		// エラーの本文にキーは含まれないが、長いので先頭だけをログ用に返す。
		return nil, fmt.Errorf("OpenAI の API がエラーを返しました: %s: %.300s", resp.Status, raw)
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("OpenAI の応答を読めません: %w", err)
	}
	if len(out.Choices) == 0 {
		return nil, errors.New("OpenAI の応答に訳がありません")
	}

	var tr map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.Choices[0].Message.Content)), &tr); err != nil {
		return nil, fmt.Errorf("訳が JSON ではありません: %w", err)
	}
	result := domain.Translations{}
	for _, lang := range langs {
		if s, ok := tr[lang].(string); ok {
			result[lang] = s
		}
	}
	return result, nil
}
