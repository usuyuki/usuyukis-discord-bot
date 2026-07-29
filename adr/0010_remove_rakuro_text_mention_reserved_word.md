# 0010. テキストの`@Rakuro`/`@rakuro`表記をコマンド起動用の予約語として扱うのをやめる

## ステータス

決定（Accepted）

## コンテキスト

[0007](./0007_dakoku_merged_into_keyword_custom_response.md)で打刻Bot固有ロジックは廃止したが、`IncomingMessage.MentionsRakuro()`（discordgoの構造化メンション`MentionsBotID`と、本文中のテキストとしての`@Rakuro`/`@rakuro`表記を統一的に判定するメソッド）自体は、keywordコマンド解釈・俳句判定の二重応答回避に引き続き使うロジックとして残していた。

この結果、`keyword_handler.go`の`HandleMessage`は`MentionsRakuro()`が真のメッセージを無条件に`handleCommand`（`keyword add/remove/list`コマンド解釈）へ振り分けていた。そのため、運用者が`@Rakuro`または`@rakuro`という文字列そのものをキーワード自動応答として登録しても、その文字列を含むメッセージは常にコマンドモードとして扱われ、`handleAutoReply`（キーワード自動応答マッチ処理）に一切到達せず、登録した応答が返らないという問題が発生した。

`@Rakuro`をコマンド起動の予約語として特別扱いする設計そのものが、キーワード自動応答の対象文字列と衝突する構造的な原因になっている。

## 決定

テキストの`@Rakuro`/`@rakuro`表記を検知する仕組み（`IncomingMessage.MentionsRakuro()`, `mentionsRakuroText()`, `isRakuroMentionToken()`, `isWordBoundaryRune()`）を削除する。Bot起動判定（keywordコマンド解釈・俳句/短歌判定の対象外判定）はdiscordgoの構造化メンション（`MentionsBotID`）のみで行う。

- `keyword_handler.go`の`HandleMessage`は`msg.MentionsBotID`のみでコマンドモードかどうかを判定する
- `parseKeywordCommand`は`@Rakuro`系トークンの除去処理を行わず、`<@123456>`等の構造化メンション文字列の除去のみ行う
- `haiku_handler.go`も同様に`msg.MentionsBotID`のみで俳句/短歌判定の対象外を判定する

## 影響

- `@Rakuro`/`@rakuro`という文字列は今後ただの文字列として扱われる。管理者はこれをキーワードとして`keyword add @Rakuro <応答>`のように登録すれば、通常のキーワード自動応答として機能する
- Botへのコマンド発行は構造化メンション（Discord上で実際にBotをメンションする操作）でのみ行える。テキストとして`@Rakuro`と打っただけではコマンドとして解釈されなくなる
- README.mdの「MESSAGE CONTENT INTENTが`@Rakuro`表記によるコマンド解釈に必須」という記述を削除する（構造化メンションの検知自体は`MentionsBotID`がdiscordgoから渡されるため、引き続きMESSAGE CONTENT INTENTは俳句/短歌判定やキーワード自動応答の本文解析に必要）
