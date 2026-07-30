# usuyukis-discord-bot

複数のDiscordサーバー（ギルド）を横断的に管理できるBot基盤。単一のBotトークン・単一プロセスで複数ギルドに参加し、以下の機能を提供する。

| 機能 | 内容 |
|---|---|
| キーワード自動応答 | Botへの構造化メンション（`@Bot keyword add <キーワード> <応答>`。`@Bot`部分は実際にDiscord上でBotをメンションする操作を指す）で登録したキーワードが発言に含まれていたら自動応答する（例: ぬるぽ→ガッ）。同じキーワードに複数の応答を登録でき、その場合はマッチ時にランダムで1つ選ばれる。`keyword remove <キーワード> <応答>` で特定の応答のみ削除、`keyword remove <キーワード>`（応答省略）でキーワードごと削除。応答文言に `{$now}` と入れるとマッチ時の現在日時（JST）に展開されるため、打刻Bot的な用途（例: 時刻を尋ねるキーワード → `今は{$now}だよ`）もこの仕組みで登録する |
| 俳句・短歌検知 | 形態素解析（[kagome](https://github.com/ikawaha/kagome)）で投稿の拍数を数え、5-7-5（俳句）または5-7-5-7-7（短歌）と判定できたら句読点区切りで**投稿元チャンネルへ**通知する（他チャンネルへは通知しない） |
| 絵文字追加通知 | ギルドへ絵文字が追加されたことを検知して通知する |
| スロット | `@Bot slot` とメンションすると、ギルドのカスタム絵文字（3個未満の場合は標準絵文字セット）から3つを独立にランダム抽選する。3つ揃うと大当たり、2つ揃うと小当たり |
| ヘルプ表示 | `@Bot help` または `@Bot usage` とメンションすると、この機能一覧をチャット上に返信する |
| 管理画面 | キーワード・通知先チャンネルをブラウザから編集できる簡易CMS（認証なし・localhost限定） |
| devモード | `DEV_MODE=true` 設定時、`DEV_CHANNEL_ID` で指定したチャンネル以外の投稿には全機能が反応しなくなる。絵文字追加通知はチャンネルを持たないため、dev mode中は全ギルドで通知自体を停止する（動作確認時の誤爆防止） |

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
  domain/                        … 値オブジェクト・純粋ロジック（keyword, notifychannel, haiku）
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
   - `MESSAGE CONTENT INTENT`（キーワード検知・俳句/短歌判定に必須）
4. 「OAuth2 > URL Generator」で `bot` スコープと以下の権限を選択し、生成されたURLからサーバーへ招待
   - `Send Messages`, `Read Message History`, `View Channels`

### 2. 環境変数の設定

```
cp .env.example .env
```

`.env` を編集し、`DISCORD_BOT_TOKEN` と `POSTGRES_PASSWORD` を設定する。

動作確認用サーバーなどで特定チャンネル以外への誤爆を防ぎたい場合は `DEV_MODE=true` と `DEV_CHANNEL_ID`（反応させたいチャンネルのID）を設定する。dev mode有効時は指定チャンネル以外の投稿にBotの全ハンドラが一切反応しなくなる。絵文字追加通知はチャンネルを持たないイベントのため、dev mode有効時は全ギルドで通知が停止する。

### 3. 起動

`compose.yml` の `bot` サービスは `image`（GHCRのビルド済みイメージ）と `build`（ローカルDockerfile）の両方を指定している。用途に応じて起動方法を使い分ける。

**本番運用（GHCRのビルド済みイメージを使う）**

`main` ブランチへのpush時にCIが `ghcr.io/usuyuki/usuyukis-discord-bot` へマルチアーキテクチャ（`linux/amd64`/`linux/arm64`）イメージを自動publishしている（詳細は [adr/0006_ci_build_pipeline.md](./adr/0006_ci_build_pipeline.md)）。ローカルビルドせずこのイメージをpullして使う。

```
docker compose pull bot
docker compose up -d
```

イメージタグは `.env` の `BOT_IMAGE_TAG`（未設定時は `latest`）で切り替えられる。

**開発（ローカルのソースからビルドする）**

```
docker compose up -d --build
```

`--build` を付けるとローカルの `Dockerfile` からビルドしたイメージで起動する（`BOT_IMAGE_TAG` で指定したタグに上書きされるため、以後 `docker compose up -d`（`--build`なし）だけだとローカルビルド版が使われ続ける点に注意。GHCR版に戻す場合は `docker compose pull bot` を再実行する）。

初回起動時に `migrations/` のマイグレーションが自動適用される。

### 4. 管理画面へのアクセス

`http://localhost:8080`（`ADMIN_PORT` で変更可能）。Dockerホスト機自身に加え、同一LAN内の他端末からも `http://<ホストのIP>:8080` でアクセスできる。**認証を実装していないため、自宅サーバー等の信頼できるネットワーク内でのみ利用し、インターネットに公開しないこと。** `docker-compose.yml` はホストの全ネットワークインターフェースへポートをバインドしているため、ルーターのポート開放等でインターネット側から到達できる状態にしないよう注意すること（詳細は [adr/0003_admin_server_no_auth.md](./adr/0003_admin_server_no_auth.md)）。

## 開発

### コマンド一覧

| コマンド | 内容 |
|---|---|
| `make test` | `go test ./...` を実行する |
| `make test-verbose` | `-v` 付きでテストを実行する |
| `make test-coverage` | カバレッジを計測し `coverage.html` を出力する |
| `make format` | `go fmt` とgolangci-lintの `--fix` でコードを整形する |
| `make lint` | `format` を実行した上でgolangci-lintの静的解析を行う |
| `make vet` | `go vet ./...` を実行する |
| `make check` | `lint` → `vet` → `test` をまとめて実行する（CI相当） |

golangci-lintは `go.mod` の `tool` ディレクティブで管理しているため、追加のインストール作業なしに `go tool golangci-lint` として実行できる。設定内容は [`.golangci.yml`](./.golangci.yml) を参照。

### CI

`.github/workflows/` にGitHub Actionsのワークフローを配置している。詳細な設計判断は [adr/0006_ci_build_pipeline.md](./adr/0006_ci_build_pipeline.md) を参照。

| ワークフロー | トリガー | 内容 |
|---|---|---|
| `000_labeler.yml` | PR作成/更新時 | 変更ファイルに応じてPRへ自動でラベルを付与する |
| `100_test.yml` | PR作成/更新時（`*.go`, `go.mod`, `go.sum` 変更時） | `go test -v ./...` とカバレッジ計測、Codecovへのアップロード |
| `101_lint.yml` | PR作成/更新時（`*.go`, `go.mod`, `go.sum`, `.golangci.yml` 変更時） | `go vet`、`gofmt` の整形チェック、`go tool golangci-lint run` |
| `900_build.yml` | mainブランチへのpush時 | `linux/amd64`/`linux/arm64` のマルチアーキテクチャDockerイメージをビルドし `ghcr.io/usuyuki/usuyukis-discord-bot` へpushする。結果を `NOTIFY_DISCORD_WEBHOOK` (GitHub Secrets) 宛にDiscord通知する |

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
