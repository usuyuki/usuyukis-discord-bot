# 0002. Webhook受信機能をスコープから除外

## ステータス

決定（Accepted）

## コンテキスト

[0001](0001_initial_architecture.md) では外部サービスからHTTPでメッセージを受け取りDiscordチャンネルへ投稿するWebhook受信機能（`internal/usecase/webhookpost`, `internal/interface/webhook`）を実装対象に含めていた。実装着手前にユーザーから「webhookのやつはなくてもいいや」と不要である旨の申告があった。

## 決定

Webhook受信機能を実装対象から除外する。

- `internal/usecase/webhookpost` を削除
- `internal/interface/webhook` を作成しない
- `internal/config.Config` から `WebhookPort`, `WebhookSharedSecret` を削除
- `docker-compose.yml` にWebhook用のポート公開設定を追加しない

## 影響

- 実装対象は「打刻Bot」「キーワード自動応答」「俳句検知（形態素解析）」「絵文字追加通知」「管理Webサイト」の5機能となる
- `MessageSender` port（`usecase/haiku`, `usecase/emoji` が定義）は引き続き必要なため、`infrastructure/discord/message_sender.go` は維持する
- 将来Webhook受信が必要になった場合は、[0001](0001_initial_architecture.md) のレイヤー構成にそのまま従い `internal/usecase/webhookpost` と `internal/interface/webhook` を追加すればよい（アーキテクチャ自体への影響はない）
