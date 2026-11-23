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
| Node.js              | 24（LTS）  |
| Next.js              | 16.4       |
| React                | 19.3       |
| TypeScript           | 5.9        |
| Tailwind CSS         | 4.3        |

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

### フロントエンドの設計方針

- `app/` は URL に対応するページを置き、API 通信、データ取得・更新、画面の状態管理を担当します。API を呼び出すのは `app/` の中だけです。
- `components/` は表示に専念し、表示データとイベント callback を props で受け取ります。API クライアントやページのデータ管理を import しません。
- `components/ticket/` や `components/map/` には機能ごとの UI、`components/ui/` には shadcn/ui の共通 UI を置きます。共通 UI を直接変更せず、必要に応じて機能コンポーネントで包みます。
- `lib/api/` には OpenAPI から生成した型と API クライアントを置きます。API の変更は `openapi/openapi.yaml` を起点にします。
- 画面文言は `locales/` の辞書を使います。Next.js は静的エクスポートするため、API Routes、Server Actions、SSR は使いません。

shadcn/ui のコンポーネントは `frontend/` で `npx shadcn@latest add <component>`（例: `npx shadcn@latest add button`）を実行して追加します。生成先は `src/components/ui/` です。

shadcn が `import { cn } from "cn"` で生成することがあるので、追加したら `@/lib/utils` に直し、`package.json` に `cn` が入っていたら `npm uninstall cn` で外してください。

<p align="right">(<a href="#top">トップへ</a>)</p>

## 開発環境構築

環境構築の Issue（#2、#4）が終わったら、ここに手順を書きます。書く内容は次のとおりです。

### コンテナの作成と起動

Docker Desktop を入れて起動しておきます（WSL を使う人は、Docker Desktop の Settings → Resources → WSL integration で自分のディストリビューションを有効にします）。リポジトリのルートで次を実行します。

```sh
docker compose up --build
```

今はフロントエンドだけが起動します。バックエンドは #7 で追加します。

### 動作確認

ブラウザで http://localhost:3000 を開き、「アトクル」と表示されれば OK です。`frontend/src` のファイルを保存すると、画面に反映されます。

### 環境変数の一覧

| 変数名 | 役割 | デフォルト値 | DEV 環境での値 |
| ------ | ---- | ------------ | -------------- |
| `TURSO_DATABASE_URL` | DB の URL（`libsql://...`、手元で試すときは `file:local.db`） | なし | Turso の DB の URL |
| `TURSO_AUTH_TOKEN` | Turso の接続トークン（`file:` のときは空） | なし | `turso db tokens create` で発行したもの |
| `PORT` | API サーバーのポート | `8080` | `8080` |
| `STATIC_DIR` | FE のビルド結果（`frontend/out/`）の場所。設定すると `/api` 以外の URL で FE のページを返す | なし（API だけ動かす） | 空（`next dev` を使う） |
| `CORS_ALLOW_ORIGINS` | 別のオリジンから API を呼ぶことを許可する（カンマ区切り）。本番は同じオリジンなので空 | なし | `http://localhost:3000` |
| `VAPID_PUBLIC_KEY` | 通知（Web Push）の VAPID 公開鍵。3つそろっていないと通知を使わずに動く | なし（通知を使わない） | 空 |
| `VAPID_PRIVATE_KEY` | 通知の VAPID 秘密鍵。Git に上げない。作り直すとそれまでの購読が使えなくなる | なし | 空 |
| `VAPID_SUBJECT` | 通知を送る人の連絡先（`mailto:` か `https:` の URL） | なし | 空 |
| `OPENAI_API_KEY` | 備考の定型文を訳す OpenAI の API キー。Git に上げない。`OPENAI_MODEL` とそろっていないと、定型文を訳さずに登録する | なし（訳さない） | 空 |
| `OPENAI_MODEL` | 定型文の翻訳に使う OpenAI のモデル名 | なし（訳さない） | 空 |
| `NEXT_PUBLIC_API_BASE_URL` | FE から呼ぶ API の URL（`frontend/.env.local` に書く）。本番は同じオリジンなので空 | `/api` | `http://localhost:8080/api` |

`backend/.env.example` を `backend/.env` にコピーして値を埋めてください。

DB の接続情報などの秘密情報は `.env` に書き、Git には上げないでください。

### コマンド一覧

| コマンド（`backend/` で実行） | 実行する処理 |
| -------- | ------------ |
| `set -a && . ./.env && set +a && go run ./cmd/server` | `.env` の環境変数を読み込んで API サーバーを起動する（ポートは `PORT`、なければ 8080）。起動時に DB のテーブルを作る（足りないものだけ足す）。`http://localhost:8080/api/health` で `{"status":"ok"}` が返れば OK |
| `go test ./...` | テストを実行する（Turso の設定がなければ、手元の SQLite ファイルだけで試す） |
| `go generate ./...` | `openapi/openapi.yaml` から `internal/gen/api.gen.go` を生成し直す |

| コマンド（`frontend/` で実行） | 実行する処理 |
| -------- | ------------ |
| `npm install` | パッケージを入れる（Docker を使わずに動かすとき。Node.js 24 が必要） |
| `npm run dev` | 開発サーバーを起動する（http://localhost:3000） |
| `npm run build` | 静的エクスポートでビルドする。`frontend/out/` に HTML・CSS・JS ができる |
| `npm run lint` | ESLint でチェックする |
| `npx shadcn@latest add <component>` | shadcn/ui のコンポーネントを `src/components/ui/` に追加する |
| `npm run gen:api` | `openapi/openapi.yaml` から API の型（`src/lib/api/schema.d.ts`）を作り直す。仕様書を変えたら実行する |

Docker でビルドだけを確かめるときは、`frontend/` で `docker build --target build .` を実行します。

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
