# 0014. チャンネル作成コマンドにリアクションによる承認投票制を導入する

## ステータス

決定（Accepted）

## コンテキスト

[ADR 0013](./0013_channel_creation_bot_command.md)でチャンネル作成をBotコマンド経由の代理作成に変更したが、コマンドを実行した瞬間に即座にチャンネルが作成される設計のままだと、誰でも際限なくチャンネルを作成できてしまい、チャンネルの乱立を招く懸念があった。

## 決定

チャンネル作成コマンドの実行結果を「即時作成」から「提案 → リアクションによる承認投票 → 可決時に作成」という2段階のフローに変更する。

- `@Bot channel create <name>` が実行されると、Botはチャンネルをまだ作成せず、実行されたチャンネルに「`#<name>` を作ろうとしています…✅で`N`人以上のリアクションが集まると作成が可決されます！」という提案メッセージを送信し、Bot自身が最初の✅リアクションを付ける（`internal/usecase/channel.UseCase.Propose`）
- 提案は`channel_create_proposals`テーブル（`internal/infrastructure/postgres/channel_proposal_repository.go`）に永続化する。Botプロセスが再起動しても未可決の提案が失われず、リアクションの積み増しに対応できるようにするため
- ユーザーが提案メッセージへリアクションするたびに（`MessageReactionAdd`イベント）、Botはそのメッセージに付いた全リアクションからBot自身を除いたユニークユーザー数を数え直す（`internal/infrastructure/discord/channel_proposal.go`の`ChannelApprovalCounter`）。絵文字の種類は問わず、任意の絵文字によるリアクションを賛成表明として扱う。同一ユーザーが複数の絵文字でリアクションしても1人としてしかカウントしない
- 必要承認人数（デフォルト2人、提案者自身のリアクションも含む）に達した時点でBotがチャンネルを作成し、提案を解決済み（`resolved = true`）としてマークする。解決済みの提案は再度リアクションが増えても二重にチャンネルを作成しない
- 必要承認人数はギルドごとに管理画面（`internal/interface/admin`のギルド詳細ページ）から設定できる（`guild_channel_create_settings`テーブル）。未設定のギルドはデフォルト値2を使う

新規追加した主なコンポーネント:

- `internal/domain/channel`: `RequiredApprovals`値オブジェクト、`IsApproved`判定関数
- `internal/usecase/channel`: `Propose`/`RecordReaction`/`GetRequiredApprovals`/`SetRequiredApprovals`
- `internal/infrastructure/discord/channel_proposal.go`: 提案メッセージ送信・リアクション集計のdiscordgo実装
- `internal/infrastructure/postgres`: `channel_proposal_repository.go`, `channel_setting_repository.go`
- `internal/interface/discordbot/reaction_handler.go`: `MessageReactionAdd`イベントを受けるハンドラ

Botのセッションには`GuildMessageReactions` Intent（非Privileged）を追加で要求する。

## 影響

- チャンネル作成には複数人の賛同が必要になり、単独ユーザーによる乱立を防げる
- 必要承認人数のカウントに提案者自身のリアクションを含めるため、実質的には「提案者 + 追加で`N-1`人」の賛同で可決される。この仕様はチームの合意形成コストと乱立防止のバランスを取ったもので、必要承認人数を1に設定すれば従来通りの即時作成に近い運用にもできる
- リアクション集計は`MessageReactionAdd`イベントのたびにDiscord APIへ問い合わせて現在の状態を数え直す設計のため、リアクション取り消し（`MessageReactionRemove`）によって一度到達した承認数が後から減ることは検知しない（可決後は`resolved`フラグでロックされるため実害はないが、可決前の取り消しは次のリアクション追加まで反映されない）
- 監査目的や誤操作からの復旧のため、提案の削除・却下を行うコマンドは本ADR時点では実装していない。可決されない提案は放置しても実害はなくデータベースに残り続けるのみ
