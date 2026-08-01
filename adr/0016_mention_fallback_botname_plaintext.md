# 0016. 構造化メンションを補う地の文@BotName表記の救済

## ステータス

決定（Accepted）。

## コンテキスト

全てのテキストコマンド（`help`/`keyword`/`channel create`等）は、Botへの構造化メンション（`<@BotID>`）が本文に含まれることを起点に判定している。

運用上、過去にBotへ送った投稿をコピー&ペーストして再送する使い方が発生した。Discordのメッセージコピーは構造化メンションのMarkdown表記をそのまま複製するとは限らず、実際には地の文の`@BotName`という平文（`m.Mentions`に追加されずBot宛メンションとして解釈されない文字列）になるケースがある。この場合、Discordクライアント上の見た目は本来の構造化メンションと区別がつかない（同じ装飾で表示される）にもかかわらず、Bot側では`MentionsBotID`が`false`のままとなり、エラーも出さず全コマンドが無反応になる。ユーザーからは「Botが壊れた」ようにしか見えず、切り分けが困難だった。

## 決定

`internal/infrastructure/discord/event_bridge.go`の`detectsMentionsBot`にて、以下のいずれかを満たす場合に`MentionsBotID`を真とする。

- 構造化メンション（`m.Mentions`）にBotのユーザーIDが含まれる（従来通り）
- 本文の**先頭単語のみ**が、`@`（半角・全角`＠`の両方を許容）を除いた上でBotのUsernameまたはGlobalName（表示名。大文字小文字を無視）のいずれかと一致する

同様に、コマンド本体を解析する`internal/interface/discordbot/handler.go`の`stripMentionTokens`も、構造化メンショントークンの除去に加えて先頭フィールドがBot名候補のいずれかと一致する場合はそれも除去するよう拡張した。ただし構造化メンションを1つでも除去した場合は、平文フォールバックによる先頭フィールドの除去は行わない（`<@botID> @someone ...`のように構造化メンション直後に平文の`@BotName`風トークンが続く入力で、ユーザーが意図した引数まで誤って消えることを防ぐため）。

誤爆防止のため、Bot名との一致判定は本文の**先頭単語のみ**を対象とする。文中の任意の位置に偶然Bot名と同じ単語が現れても反応しない。

一致判定用のBot名候補（Username/GlobalName）の組み立てと、`@`/`＠`除去＋大文字小文字無視の比較ロジックは`internal/interface/discordbot/handler.go`の`MatchesBotMentionName`に集約し、`detectsMentionsBot`（infrastructure層）と`stripMentionTokens`（interface層）の双方から参照する。判定基準を1箇所に閉じることで、将来の修正で両者の挙動がズレることを防ぐ。

## 影響

- 過去投稿のコピペ等で構造化メンションが失われた場合でも、本文が`@BotName ...`（Username・GlobalNameどちらでも、半角`@`・全角`＠`どちらでも）で始まっていればコマンドとして認識できるようになる
- Botの表示名（Username/GlobalName）が変更されると、この救済ロジックの一致対象も追従して変わる（`s.State.User`を都度参照するため設定の二重管理は発生しない）
- 誤反応リスクとして、たまたま先頭が`@BotName`から始まる無関係な発言（他Botの話題等）にも反応してしまう可能性はあるが、先頭一致に限定することで許容範囲に抑えている
- **既存機能への副作用**: `MentionsBotID`が真になる範囲が広がったことで、`MentionsBotID`が偽であることを前提に動作していた既存ハンドラの挙動が変わる。
  - `HaikuHandler`（`haiku_handler.go`）は`MentionsBotID`が真の間は俳句・短歌判定を行わない。そのため、5-7-5等の投稿がたまたま平文`@BotName`で始まっていた場合、この救済ロジック導入前は判定対象だったが、導入後は`MentionsBotID`が真になり判定がスキップされる
  - `KeywordHandler`（`keyword_handler.go`）は`MentionsBotID`が真の場合コマンド解析経路（`handleCommand`）に、偽の場合は自動応答経路（`handleAutoReply`）に振り分ける。登録済みキーワードにマッチする文言が平文`@BotName`で始まるメッセージに含まれていた場合、導入前は自動応答が発火したが、導入後はコマンド経路に振り分けられ`parseKeywordCommand`がnilを返して自動応答が発火しなくなる
  - 上記2点は許容する仕様とする（コピペ救済によるコマンド認識を優先し、平文`@BotName`から始まる投稿は原則コマンドとして扱う）。将来この分岐条件を変更する場合は本ADRを踏まえて新たなADRを起こすこと
