# 0007. 打刻Bot機能を廃止し、キーワード自動応答の{$now}プレースホルダーに統一する

## ステータス

決定（Accepted）

## コンテキスト

初期アーキテクチャ（[0001](./0001_initial_architecture.md)）では、`@Rakuro` メンションで現在時刻を固定フォーマットで返信する打刻Bot機能を、キーワード自動応答とは別の独立したユースケース（`domain/dakoku`, `usecase/dakoku`, `interface/discordbot/dakoku_handler.go`）として実装していた。

この打刻Botは応答文言・トリガー（`@Rakuro`固定）ともにハードコードされており、キーワード自動応答のように管理画面やDiscordコマンドから応答文言をカスタマイズすることができなかった。運用上、打刻以外の用途でも「現在時刻を含む応答」を登録したい要望があり、専用機能として残す理由が薄れていた。

## 決定

打刻Bot機能（`domain/dakoku`, `usecase/dakoku`, `interface/discordbot/dakoku_handler.go`とmain.goでの登録）を削除し、キーワード自動応答に一本化する。

- キーワード自動応答の応答文言に `{$now}` というプレースホルダーを埋め込めるようにし、マッチ時に現在日時（JST, `2006-01-02 15:04:05`形式）へ展開する
- 展開ロジックは `domain/keyword.Keyword.RandomResponse(now time.Time) string` に実装する。ランダム選択後の文字列に対して`strings.ReplaceAll`で置換するのみで、外部ライブラリには依存しない
- 呼び出し側（`interface/discordbot/keyword_handler.go`）は打刻Bot時代と同様に `now func() time.Time` フィールドを持ち、テストで時刻を固定できるようにする
- `@Rakuro`固定の打刻トリガーは廃止する。時刻入りの応答が欲しい場合は、管理者が任意のキーワードに対し `{$now}` を含む応答をキーワード自動応答として登録する運用に変える
- `IncomingMessage.MentionsRakuro()` 自体（構造化メンション・テキストの`@Rakuro`表記の統一判定）はkeywordコマンド解釈・俳句判定の二重応答回避に引き続き使うため残す。廃止するのはあくまで「`@Rakuro`メンションで時刻を固定応答する」打刻専用ロジックのみ

## 影響

- `internal/domain/dakoku`, `internal/usecase/dakoku`, `internal/interface/discordbot/dakoku_handler.go`（およびテスト）を削除した
- `keyword.Keyword.RandomResponse()` のシグネチャが `RandomResponse(now time.Time)` に変更された。呼び出し元は `interface/discordbot/keyword_handler.go` のみで、テストも合わせて更新した
- 既存の「`@Rakuro`と発言すれば無条件で時刻が返る」動作はなくなる。運用移行時は管理画面またはDiscordコマンド（`keyword add <トリガー> 今は{$now}だよ`のような形）で改めて登録が必要
- 応答文言に`{$now}`という4文字が偶然含まれるキーワードを登録すると意図せず日時に展開される。予約プレースホルダーであることをREADME・管理画面に明記して周知する
