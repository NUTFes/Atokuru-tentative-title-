<div id="top"></div>

## 使用技術一覧

<!-- シールド一覧 -->
<p style="display: inline">
  <!-- フロントエンドのフレームワーク一覧 -->
  <img src="https://img.shields.io/badge/-Next.js-000000.svg?logo=nextdotjs&style=for-the-badge">
  <img src="https://img.shields.io/badge/-React-20232A.svg?logo=react&style=for-the-badge">
  <img src="https://img.shields.io/badge/-Tailwind%20CSS-0F172A.svg?logo=tailwindcss&style=for-the-badge">
  <img src="https://img.shields.io/badge/-shadcn%2Fui-000000.svg?logo=shadcnui&style=for-the-badge">
  <!-- フロントエンドの言語一覧 -->
  <img src="https://img.shields.io/badge/-TypeScript-007ACC.svg?logo=typescript&style=for-the-badge&logoColor=white">
  <!-- バックエンドのフレームワーク一覧 -->
  <img src="https://img.shields.io/badge/-Echo-00ADD8.svg?style=for-the-badge">
  <!-- バックエンドの言語一覧 -->
  <img src="https://img.shields.io/badge/-Go-00ADD8.svg?logo=go&style=for-the-badge&logoColor=white">
  <!-- ミドルウェア一覧 -->
  <img src="https://img.shields.io/badge/-Turso-4FF8D2.svg?logo=turso&style=for-the-badge&logoColor=black">
  <!-- その他 -->
  <img src="https://img.shields.io/badge/-OpenAPI-6BA539.svg?logo=openapiinitiative&style=for-the-badge&logoColor=white">
</p>

## 目次

1. [プロジェクトについて](#プロジェクトについて)
2. [環境](#環境)
3. [ディレクトリ構成](#ディレクトリ構成)
4. [開発環境構築](#開発環境構築)
5. [開発の進め方](#開発の進め方)
6. [トラブルシューティング](#トラブルシューティング)

<br />

<div align="right">
  <a href="AGENTS.md"><strong>AI 向けの規約（AGENTS.md）»</strong></a>
</div>
<br />

## プロジェクト名

アトクル（仮）

## プロジェクトについて

アトクルは、学園祭の模擬店の整理券をデジタルにする Web アプリです。

模擬店では調理の都合で待機列ができ、30分近く待つこともあります。今は店ごとに紙の整理券で対応しているので、来場者は自分の注文があとどのくらいでできるのか分かりません。そのため受け取るまでその場を離れられず、学園祭を回る時間が減ってしまいます。アトクルでは、整理券をスマホで受け取り、前に何人いるかをどこにいても確かめられるようにします。

画面は、来場者が使うユーザー画面と、店が使う模擬店画面の2つです。

- ユーザー画面では、店で受け取った QR コードを読み取ると整理券が表示され、前に何人いるかと調理の状況を確かめられます。マップでは、屋台ごとの混雑度と営業状況を見られます。
- 模擬店画面では、整理券を発行し、注文を調理完了・受け取り済みにします。営業状況（営業中・準備中・売り切れ・閉店）の切り替えもここで行います。

来場者のアカウントは作りません。整理券は推測されにくいトークンで見分け、来場者のスマホのブラウザに保存します。

タスクは [Issues](https://github.com/NUTFes/Atokuru-tentative-title-/issues) で管理しています。

<p align="right">(<a href="#top">トップへ</a>)</p>

## 環境

| 言語・フレームワーク | バージョン |
| -------------------- | ---------- |
| Go                   | 1.26       |
| Echo                 | v4.16      |
| Node.js              | 未記入     |
| Next.js              | 未記入     |
| React                | 未記入     |
| TypeScript           | 未記入     |
| Tailwind CSS         | 未記入     |

| 用途     | 使うもの                                   |
| -------- | ------------------------------------------ |
| API 仕様 | OpenAPI（`openapi/openapi.yaml`）           |
| コード生成 | oapi-codegen（Go）、openapi-typescript（TS） |
| API 呼び出し | openapi-fetch                           |
| ORM      | GORM                                       |
| DB       | Turso（東京リージョン）                     |

バージョンは環境構築の Issue（#2、#4）で決まったら書き込んでください。そのほかのパッケージは `backend/go.mod` と `frontend/package.json` を見てください。

<p align="right">(<a href="#top">トップへ</a>)</p>

## ディレクトリ構成

```
.
├── openapi
│   └── openapi.yaml          フロントとバックで共通の仕様書
├── backend
│   ├── cmd
│   │   └── server            main.go（起動、DB 接続、ルーティングの登録）
│   └── internal
│       ├── gen               oapi-codegen の出力（手で編集しない）
│       ├── handler           gen のインターフェースを実装し、usecase を呼ぶ
│       ├── usecase           残り人数の計算などの判断
│       ├── repository        GORM で DB を読み書きする
│       ├── domain            entity（stall、ticket など）
│       └── infra
│           ├── db            Turso への接続
│           └── notify        Web Push の送信（後回し）
├── frontend
│   ├── public                マップの画像、manifest.json、sw.js
│   └── src
│       ├── app               ページ（URL と1対1）
│       │   ├── map
│       │   └── stall
│       │       ├── login
│       │       ├── tickets
│       │       ├── status
│       │       └── settings
│       ├── components
│       │   ├── ui            shadcn が生成するもの
│       │   ├── ticket
│       │   ├── nav
│       │   ├── sheet
│       │   └── map
│       ├── lib
│       │   └── api           openapi-fetch のクライアントと生成した型
│       └── locales           多言語の辞書（ja.json など）
├── AGENTS.md                 AI 向けの規約
├── CLAUDE.md                 AGENTS.md を読み込む
└── README.md
```

バックエンドの依存の向きは handler → usecase → repository → domain の一方向です。フロントエンドで API を呼ぶのは `app/` の中だけにして、`components/` は見た目だけを担当します。詳しくは [AGENTS.md](AGENTS.md) を見てください。

<p align="right">(<a href="#top">トップへ</a>)</p>

## 開発環境構築

環境構築の Issue（#2、#4）が終わったら、ここに手順を書きます。書く内容は次のとおりです。

### コンテナの作成と起動

`.env` の作り方と、ローカルでまとめて起動するコマンドを書きます。

### 動作確認

起動したあとに開く URL と、正しく動いているときの表示を書きます。

### 環境変数の一覧

| 変数名 | 役割 | デフォルト値 | DEV 環境での値 |
| ------ | ---- | ------------ | -------------- |
| `TURSO_DATABASE_URL` | DB の URL（`libsql://...`、手元で試すときは `file:local.db`） | なし | Turso の DB の URL |
| `TURSO_AUTH_TOKEN` | Turso の接続トークン（`file:` のときは空） | なし | `turso db tokens create` で発行したもの |
| `PORT` | API サーバーのポート | `8080` | `8080` |
| `STATIC_DIR` | FE のビルド結果（`frontend/out/`）の場所。設定すると `/api` 以外の URL で FE のページを返す | なし（API だけ動かす） | 空（`next dev` を使う） |
| `CORS_ALLOW_ORIGINS` | 別のオリジンから API を呼ぶことを許可する（カンマ区切り）。本番は同じオリジンなので空 | なし | `http://localhost:3000` |

`backend/.env.example` を `backend/.env` にコピーして値を埋めてください。

DB の接続情報などの秘密情報は `.env` に書き、Git には上げないでください。

### コマンド一覧

| コマンド（`backend/` で実行） | 実行する処理 |
| -------- | ------------ |
| `set -a && . ./.env && set +a && go run ./cmd/server` | `.env` の環境変数を読み込んで API サーバーを起動する（ポートは `PORT`、なければ 8080）。起動時に DB のテーブルを作る（足りないものだけ足す）。`http://localhost:8080/api/health` で `{"status":"ok"}` が返れば OK |
| `go test ./...` | テストを実行する（Turso の設定がなければ、手元の SQLite ファイルだけで試す） |
| `go generate ./...` | `openapi/openapi.yaml` から `internal/gen/api.gen.go` を生成し直す |

<p align="right">(<a href="#top">トップへ</a>)</p>

## 開発の進め方

1. [Issues](https://github.com/NUTFes/Atokuru-tentative-title-/issues) から担当の Issue を選びます。タイトルの番号が、やる順番です。
2. Issue ごとに `種類/名前/Issue番号/内容` の名前でブランチを切ります（例：`feat/goldfish/19/get-ticket`）。種類は `feat`・`fix`・`docs` のどれかです。詳しくは [AGENTS.md](AGENTS.md) を見てください。
3. Issue のチェックボックスを埋めながら開発します。
4. 全部終わったら PR を出し、説明に `close #<Issue番号>` を書きます。
5. 1人にレビューしてもらってから main にマージします。main には直接 push しません。

API を変えるときは、先に `openapi/openapi.yaml` を直してから、コードを生成し直して実装します。

<p align="right">(<a href="#top">トップへ</a>)</p>

## トラブルシューティング

開発中に起きたエラーと直し方を、エラー文を見出しにしてここに書き足してください。

<p align="right">(<a href="#top">トップへ</a>)</p>
