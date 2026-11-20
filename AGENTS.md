# AGENTS.md

このリポジトリで AI（Claude Code、Codex、ChatGPT など）にコードを書かせるときの規約です。
人間もこれに従います。ChatGPT（Web 版）を使う場合は、このファイルの中身をプロンプトの先頭に貼ってください。

## アプリの概要

アトクルは、学園祭の模擬店の整理券をデジタルで管理する Web アプリです。

- **ユーザー画面**（来場者）：QR コードを読み取って整理券を受け取り、自分の前に何人いるかを確認する。マップで屋台の混雑度を見る
- **模擬店画面**（店）：整理券を発行し、調理完了・受け取り済みにする。営業状況（営業中・準備中・売り切れ・閉店）を切り替える

来場者のアカウントは作りません。整理券は推測されにくい `public_token` で識別し、端末の localStorage に保存します。

## 技術スタック

| 層 | 使うもの |
| --- | --- |
| フロント | Next.js（静的エクスポート）+ React + TypeScript + Tailwind CSS + shadcn/ui |
| API 呼び出し | openapi-typescript + openapi-fetch |
| API 仕様 | OpenAPI（`openapi/openapi.yaml`） |
| バック | Go + Echo、oapi-codegen、GORM |
| DB | Turso（東京） |

ここにないライブラリを追加するときは、PR の説明に理由を書いてください。

## ディレクトリ構成

```
openapi/openapi.yaml      フロントとバックの共通の仕様書（唯一の正）
backend/
  cmd/server/main.go      起動、DB 接続、ルーティングの登録だけ
  internal/gen/           oapi-codegen の出力
  internal/handler/       gen のインターフェースを実装。usecase を呼ぶだけ
  internal/usecase/       残り人数の計算などの判断
  internal/repository/    GORM で DB を読み書き
  internal/domain/        entity（stall.go、ticket.go など）
  internal/infra/db/      Turso への接続
  internal/infra/notify/  Web Push の送信（今は後回し）
frontend/src/
  app/                    ページ（URL と1対1）
    page.tsx              ユーザー画面のホーム（整理券一覧）
    map/                  マップ
    stall/login|tickets|status|settings/   模擬店画面
  components/ui/          shadcn が生成するもの
  components/ticket|nav|sheet|map/         自作のコンポーネント
  lib/api/                openapi-fetch のクライアントと生成した型
  lib/storage.ts          localStorage の読み書き
  locales/                ja.json、en.json などの辞書
frontend/public/          マップの画像、manifest.json、sw.js
```

新しいファイルは、この構成のどこかに置いてください。新しいフォルダを作る前に、既存のフォルダで足りないかを確認してください。

## 必ず守ること

### 共通

- **API の変更は、必ず `openapi/openapi.yaml` から始める**：仕様書を直す → コードを生成し直す → 実装する、の順番です。仕様書にない API を作ったり、生成されたコードと違う形の JSON を返したりしないでください
- **生成されたコードは手で編集しない**：`backend/internal/gen/` と `frontend/src/lib/api/` の生成ファイルは、仕様書を直して生成し直します
- **秘密情報をコミットしない**：DB の接続情報、API キーなどは `.env` に書き、Git に上げません
- **Issue の範囲だけを変更する**：頼まれていないファイルの整形やリファクタリングはしないでください

### バックエンド

- 依存の向きは **handler → usecase → repository → domain** の一方向です。逆向きの import はしません
- handler には判断を書きません。リクエストを受け取って usecase を呼び、結果を返すだけです
- DB を触るのは repository だけです。usecase から GORM を直接呼びません
- 一覧を取る API では全件を返さず、status などで絞り込みます。整理券は10万件規模になる想定です

### フロントエンド

- **API を呼ぶのは `app/` の中だけ**です。`components/` から `lib/api` を import しないでください
- コンポーネントは見た目だけを担当します。表示するデータとボタンの callback は、引数（props）で受け取ります
- `components/ui/` は shadcn が生成したものなので、基本的に編集しません。見た目を変えたいときは、自作のコンポーネントで包みます
- 静的エクスポート（`output: 'export'`）なので、**API Routes、Server Actions、SSR は使えません**。サーバーの処理はすべて Go 側に書きます
- 動的な URL（`/tickets/[id]` など）は使わず、`/ticket?id=...` のようにクエリパラメータで渡します
- データを取りに行くのは、ページを開いたときとアプリに戻ってきたとき（`visibilitychange`）だけです。定期的なポーリングはしません
- 画面の文言は、直接書かずに `locales/` の辞書から読みます（多言語対応のため）

## 決まっている値

| 項目 | 値 |
| --- | --- |
| tickets.status | `waiting`（調理中）/ `ready`（調理完了）/ `received`（受け取り済み） |
| stalls.business_status | `open`（営業中）/ `preparing`（準備中）/ `sold_out`（売り切れ）/ `closed`（閉店） |
| 混雑度 | 赤・黄・緑・灰（営業していない店は灰） |
| 前に何人 | 自分の `order_number` − 同じ屋台で `received` の最大 `order_number` − 1 |

整理券の ID は連番なので、URL や QR コードには `id` ではなく `public_token` を使います。

## Git の運用

- **main に直接 push しない**：Issue ごとにブランチを切り、PR を出して1人に見てもらってからマージします
- ブランチ名：`feature/<Issue番号>-<内容>`（例：`feature/19-get-ticket`）
- PR の説明に `close #<Issue番号>` を書いて、Issue と紐づけます
- コミットメッセージは日本語で、何をしたかが分かるように書きます
