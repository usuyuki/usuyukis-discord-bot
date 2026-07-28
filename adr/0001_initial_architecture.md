# 0001. 初期アーキテクチャの決定

## ステータス

決定（Accepted）

## コンテキスト

複数のDiscordサーバー（Discord API上の正式名称は Guild）を横断的に管理するBot基盤をGo言語で新規構築する。単一のBotトークン・単一プロセスで複数ギルドに参加し、以下を最初のユースケースとして実装する。

- `@Rakuro` メンションで現在時刻を返す打刻Bot
- キーワードをメンション経由で事前登録し、発言に含まれていたら自動応答する機能（例: ぬるぽ→ガッ）
- 5-7-5（俳句）の投稿を形態素解析による拍数判定で検知して通知する機能
- ギルドへの絵文字追加を検知して通知する機能
- 外部サービスからHTTPで受け取ったメッセージをDiscordチャンネルへ投稿するWebhook受信機能
- キーワード・通知先チャンネルをブラウザから編集できる簡易管理画面（自宅サーバーでlocalhostのみ公開・認証なし）

今後もBot機能が継続的に追加される前提のため、新機能追加が既存コードの改変を最小限に抑えて成立する構成が要件となる。

## 決定

### 1. クリーンアーキテクチャを採用し、依存方向を一方向（外側→内側）に固定する

```
interface (最外層: discordgoイベント変換, HTTPハンドラ, html/template)
    ↓ depends on
infrastructure (discordgo実装, pgx実装, kagome実装 — usecaseが定義したportを実装)
    ↓ implements ports defined in usecase
usecase (各Bot機能のアプリケーションロジック。domainのみに依存)
    ↓ depends on
domain (最内層・外部依存ゼロ。値オブジェクト・純粋ロジック)
```

- `domain` は外部ライブラリに一切依存しない純粋なGoコード
- `usecase` は `domain` のみに依存し、外部技術（discordgo, pgx, kagome, net/http）には依存しない。必要な外部機能は自分でport interfaceを定義し、`infrastructure` がそれを実装する（依存性逆転の原則）
- `infrastructure` はusecaseが定義したport interfaceの実装を担う
- `interface` は一番外側。discordgoのイベントやHTTPリクエストを受けてusecaseを呼び出す

### 2. ハンドラのプラグイン登録パターンで機能追加コストを固定する

新機能追加の手順を以下に固定する。

1. `internal/domain/<feature>/` に純粋ロジックを書く（あれば）
2. `internal/usecase/<feature>/` にユースケースとport interfaceを書く
3. 必要なら `internal/infrastructure/` にport実装を追加（既存実装を再利用できることが多い）
4. `internal/interface/discordbot/<feature>_handler.go` に `MessageHandler` 等のインターフェースを実装
5. `cmd/bot/main.go` の `router.Register(...)` に1行追加

`MessageHandler` / `EmojiUpdateHandler` インターフェースを `internal/interface/discordbot/handler.go` で契約として固定し、router は登録済み全ハンドラへイベントをブロードキャストするだけの薄い実装とする。

### 3. 俳句判定は形態素解析ベースで行う

`github.com/ikawaha/kagome/v2` + `github.com/ikawaha/kagome-dict/ipa`（純Go実装、cgo不要）を採用。投稿本文を形態素解析し、各形態素の読みをモーラ（拍）単位に分解した上で5-7-5に区切れるかを判定する。kagome依存は `infrastructure/morph` に閉じ込め、`usecase/haiku` は `MorphAnalyzer` interface越しに「モーラ列」を受け取るのみとする。

### 4. 永続化はPostgreSQL、docker-composeで管理

`keywords`, `guild_notify_channels` の2テーブルを `golang-migrate` で管理。docker-composeで `bot` + `postgres` を管理する。

### 5. 管理画面はGoサーバーサイドレンダリング、認証なし・localhost限定

`net/http` + `html/template` のみ（追加ライブラリなし）。自宅サーバーでの運用を前提に認証を設けず、docker-composeで `127.0.0.1` のみへバインドする。Discordコマンド経由の操作とusecase層を共有し、ロジックの二重実装を避ける。

## 影響

- 新機能を追加する開発者は、上記5ステップの型に従うことで既存コードへの影響を最小化できる
- `domain`/`usecase` は外部ライブラリ非依存のため、テストが高速でモック不要（フェイク実装で足りる）
- レイヤー間の依存方向は `go list -deps` 等で機械的に検証可能（README記載のコマンドで手動確認）
- kagome辞書を埋め込むため、Dockerイメージサイズは辞書データ分増加するが、外部プロセス（MeCab等）へのcgo依存は回避できる

## 運用ルール

- 以後、アーキテクチャ上の重要な決定を行った際は `adr/NNNN_<slug>.md` として本ディレクトリに追記する（連番はゼロ埋め4桁）
- `CLAUDE.md` と `README.md` は変更のたびに追従して更新する
