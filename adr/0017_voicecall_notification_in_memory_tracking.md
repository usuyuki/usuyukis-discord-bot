# 0017. ボイスチャンネル通話開始/終了通知はGuildVoiceStates Intentとプロセスローカルな在室者数カウンタで実現する

## ステータス

決定（Accepted）

## コンテキスト

ボイスチャンネルの在室者が0人→1人以上（通話開始）、1人以上→0人（通話終了）に変化したタイミングをテキストチャンネルへ通知する機能を追加する。Discordのゲートウェイから`VoiceStateUpdate`イベントを受け取るには`GuildVoiceStates` Intentが必要で、これは`internal/infrastructure/discord/session.go`が要求する既存のIntents（Guilds, GuildMessages, GuildMessageContent, GuildEmojis, GuildMessageReactions）に追加する新しい特権外Intentである。

`VoiceStateUpdate`イベント単体には「更新前後のチャンネル人数」が含まれておらず、discordgoの`Session.State`はハンドラ呼び出し時点で既に更新後の状態になっているため「更新前の人数」を復元できない。そのため、絵文字追加通知（`previousEmojis`、[0001](./0001_initial_architecture.md)時点から存在するプロセスローカルmapによる差分検出）と同様の手法として、ギルド×チャンネルごとの在室者数をプロセスローカルなカウンタ（`voiceCallTracker`）で自前管理する方針を採る。

この方式には次のトレードオフがある。

- 通話開始時刻（`startedAt`）はプロセスメモリにのみ保持される。`900_build.yml`はmainへのpushのたびに新しいDockerイメージをビルド・デプロイする（[0006](./0006_ci_build_pipeline.md)）ため、通話が進行中にBotが再起動・再デプロイされると、その通話の開始時刻情報は失われる。再起動後に`GuildCreate`が再送されて`voiceCallTracker.initGuild`が呼ばれた際は、その時点を開始時刻とみなして初期化するため、その通話が実際に終了した際の通話時間表示は「不明」ではなく「再起動時点からの経過時間」という不正確な値になる（実装上は`DurationKnown=true`だが実際の通話時間より短い値が出る）
- 複数のBotプロセスをスケールアウトさせる場合、在室者数カウンタがプロセス間で共有されないため、この機能はシングルプロセス運用が前提となる

これらは通話時間表示という補助的な情報の精度に留まる問題であり、通話開始/終了の検知自体（在室者数が0を跨いだかどうか）はプロセス起動中は正しく機能する。DBへ永続化する設計も検討したが、通話時間はログ的価値が主で、正確性のために書き込みコストと実装複雑度を増やす価値は現時点でないと判断した。

## 決定

- `internal/infrastructure/discord/session.go`のIntentsに`discordgo.IntentsGuildVoiceStates`を追加する
- 通話開始時刻はDBへ永続化せず、`voiceCallTracker`（`internal/infrastructure/discord/voicecall_tracker.go`）のプロセスローカルなmapにのみ保持する
- Bot再起動をまたぐ通話の時間表示は不正確になりうることを許容する

## 影響

- Botの招待URLに`GuildVoiceStates` Intentの許可が必要になる（特権Intentではないため、Discord Developer Portalでの追加申請は不要）
- 通話中にデプロイが走ると、その通話の終了時に表示される通話時間が実際より短くなる場合がある。正確な通話時間が必要になった場合は、`startedAt`をDBに永続化する変更を別ADRとして起こすこと
- 複数プロセスでのスケールアウト運用はこの機能単体では未対応
