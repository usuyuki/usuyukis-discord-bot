# usuyukis-discord-bot

複数のDiscordサーバー（ギルド）を横断的に管理できるBot基盤。単一のBotトークン・単一プロセスで複数ギルドに参加し、以下の機能を提供する。

| 機能 | 内容 |
|---|---|
| 打刻Bot | 発言本文に `@Rakuro`（大文字小文字を区別しない）という文字列が含まれると現在時刻を返信する |
| キーワード自動応答 | `@Bot keyword add <キーワード> <応答>` で登録したキーワードが発言に含まれていたら自動応答する（例: ぬるぽ→ガッ） |
| 俳句・短歌検知 | 形態素解析（[kagome](https://github.com/ikawaha/kagome)）で投稿の拍数を数え、5-7-5（俳句）または5-7-5-7-7（短歌）と判定できたら句読点区切りで通知する |
| 絵文字追加通知 | ギルドへ絵文字が追加されたことを検知して通知する |
| 管理画面 | キーワード・通知先チャンネルをブラウザから編集できる簡易CMS（認証なし・localhost限定） |
| devモード | `DEV_MODE=true` 設定時、`DEV_CHANNEL_ID` で指定したチャンネル以外の投稿には全機能が反応しなくなる（動作確認時の誤爆防止） |

アーキテクチャ上の意思決定は [`adr/`](./adr) ディレクトリのADR（Architecture Decision Record）を参照。特に [0001_initial_architecture.md](./adr/0001_initial_architecture.md) にレイヤー構成と拡張方法をまとめている。

## アーキテクチャ概要

クリーンアーキテクチャを採用し、依存方向は常に `interface/infrastructure → usecase → domain` の一方向。

```
interface (discordgoイベント変換, HTTPハンドラ, html/template)
    ↓
infrastructure (discordgo実装, pgx実装, kagome実装)
    ↓ implements ports defined in usecase
usecase (各Bot機能のアプリケーションロジック。domainのみに依存)
    ↓
domain (外部依存ゼロの値オブジェクト・純粋ロジック)
```

新しいBot機能を追加する手順は固定されている（詳細はADR参照）。

1. `internal/domain/<feature>/` に純粋ロジックを書く（あれば）
2. `internal/usecase/<feature>/` にユースケースとport interfaceを書く
3. 必要なら `internal/infrastructure/` にport実装を追加
4. `internal/interface/discordbot/<feature>_handler.go` に `MessageHandler` 等を実装
5. `cmd/bot/main.go` の `router.Register...(...)` に1行追加

## ディレクトリ構成

```
cmd/bot/main.go                  … エントリポイント（DI結線）
internal/
  domain/                        … 値オブジェクト・純粋ロジック（keyword, notifychannel, haiku, dakoku）
  usecase/                       … アプリケーションロジック + port interface定義
  infrastructure/
    discord/                     … discordgoセッション・MessageSender・GuildCache実装
    postgres/                    … pgxによるRepository実装、マイグレーション実行
    morph/                       … kagomeによる形態素解析実装
  interface/
    discordbot/                  … discordgoイベント→usecase変換、ハンドラのプラグイン登録
    admin/                       … 管理画面（net/http + html/template、認証なし・localhost限定）
  config/                        … 環境変数読み込み
migrations/                      … golang-migrate用SQLマイグレーション
adr/                             … Architecture Decision Record
```

## セットアップ

### 1. Discord Bot の作成

1. [Discord Developer Portal](https://discord.com/developers/applications) で新規アプリケーションを作成
2. 「Bot」タブでBotを追加し、トークンを発行（`.env` の `DISCORD_BOT_TOKEN` に設定）
3. 「Bot」タブの **Privileged Gateway Intents** で以下を有効化
   - `MESSAGE CONTENT INTENT`（キーワード検知・俳句/短歌判定・打刻Botの本文メンション検知に必須）
4. 「OAuth2 > URL Generator」で `bot` スコープと以下の権限を選択し、生成されたURLからサーバーへ招待
   - `Send Messages`, `Read Message History`, `View Channels`

### 2. 環境変数の設定

```
cp .env.example .env
```

`.env` を編集し、`DISCORD_BOT_TOKEN` と `POSTGRES_PASSWORD` を設定する。

動作確認用サーバーなどで特定チャンネル以外への誤爆を防ぎたい場合は `DEV_MODE=true` と `DEV_CHANNEL_ID`（反応させたいチャンネルのID）を設定する。dev mode有効時は指定チャンネル以外の投稿にBotの全ハンドラが一切反応しなくなる。

### 3. 起動

```
docker compose up -d --build
```

初回起動時に `migrations/` のマイグレーションが自動適用される。

### 4. 管理画面へのアクセス

`http://localhost:8080`（`ADMIN_PORT` で変更可能）。Dockerホスト機自身に加え、同一LAN内の他端末からも `http://<ホストのIP>:8080` でアクセスできる。**認証を実装していないため、自宅サーバー等の信頼できるネットワーク内でのみ利用し、インターネットに公開しないこと。** `docker-compose.yml` はホストの全ネットワークインターフェースへポートをバインドしているため、ルーターのポート開放等でインターネット側から到達できる状態にしないよう注意すること（詳細は [adr/0003_admin_server_no_auth.md](./adr/0003_admin_server_no_auth.md)）。

## 開発

### テスト実行

```
go test ./...
```

### 静的解析

```
go vet ./... && gofmt -l .
```

### レイヤー依存方向の確認

`domain` と `usecase` パッケージが外部ライブラリ（discordgo, pgx, kagome, net/http）に依存していないことを確認する。

```
go list -deps ./internal/domain/... ./internal/usecase/... | grep -E 'discordgo|jackc/pgx|ikawaha/kagome|net/http'
```

何も出力されなければ依存方向は健全。

### ローカルでのBot起動確認（Docker不使用）

```
docker compose up -d postgres
go run ./cmd/bot
```
