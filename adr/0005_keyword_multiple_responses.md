# 0005. キーワード自動応答は1キーワードにつき複数応答を持てるようにする

## ステータス

決定（Accepted）

## コンテキスト

これまでキーワード自動応答は `keywords` テーブルに `guild_id, keyword, response` を1行として保持し、`UNIQUE(guild_id, keyword)` 制約のもとで1キーワードにつき1応答しか登録できなかった。同一キーワードへの再登録は応答文言の上書き（UPSERT）として扱われていた。

Slackのカスタム応答のように、1つのキーワードに複数の応答候補を持たせ、マッチ時にその中からランダムに1つを選んで返す挙動が求められた。

## 決定

キーワードと応答を別テーブルに正規化し、1キーワード:N応答の構造に変更する。

- `keywords(id, guild_id, keyword)` … キーワード本体。`UNIQUE(guild_id, keyword)` は維持
- `keyword_responses(id, keyword_id, response)` … 応答候補。`keywords.id` への外部キーで `ON DELETE CASCADE`、`UNIQUE(keyword_id, response)` で同一応答の重複登録を防ぐ
- マイグレーションは `migrations/0002_keyword_multi_response.up.sql` で追加し、既存の `keywords.response` 列のデータは移行後に `keyword_responses` へ複製してから列を削除する

ドメイン層 `domain/keyword.Keyword` は `Response string` を `Responses []string` に変更し、`RandomResponse()` で応答候補からランダムに1件選択するロジックを持つ。

usecase層のRepository portは以下の操作に分割した。

- `AddResponse` … 指定キーワードに応答を1件追加（積み増し）。Discordの `keyword add <word> <response>` コマンドはこれを呼ぶたびに応答候補が増える
- `RemoveResponse` … 特定の応答候補のみ削除。最後の1件を消すとキーワード自体も削除される。Discordの `keyword remove <word> <response>` に対応
- `RemoveKeyword` … キーワードを応答候補ごと全削除。Discordの `keyword remove <word>`（応答省略時）に対応
- `ReplaceResponses` … 応答候補を丸ごと置き換える。Web管理画面での改行区切りテキストエリアによる一括編集に対応

Web管理画面（`internal/interface/admin`）は1キーワード1行のまま、応答列を改行区切りのテキストエリアで表示・一括編集する形にした。

## 影響

- Discordコマンドの `keyword remove` は引数の数で「特定応答の削除」と「キーワード全体削除」を切り替える仕様になった。既存の「単語のみ指定で削除」という使用感は維持しつつ、応答を追加指定すると個別削除になる
- `keyword add` の重複実行は上書きではなく応答候補の積み増しになるため、誤って同じキーワードに同じ応答を登録すると `UNIQUE(keyword_id, response)` により無視される（エラーにはしない）
- マッチング時の応答選択に `math/rand/v2` を用いるため、同一メッセージへの応答が実行のたびに変わりうる。テストで結果を固定したい場合はランダム性を考慮したアサーション（複数回試行して集合を見るなど）が必要
