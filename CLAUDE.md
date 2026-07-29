# CLAUDE.md

このリポジトリで作業する際の開発規約。

**⚠️ 注意: このリポジトリはパブリック（公開）リポジトリです。APIキー、トークン、パスワード、特定のユーザーIDなどの機密情報をソースコードやドキュメントに含めないように注意してください。**

## アーキテクチャ

クリーンアーキテクチャを採用している。依存方向は常に一方向。

```
interface → infrastructure → usecase → domain
```

- `domain`: 外部ライブラリに一切依存しない純粋なGo（値オブジェクト・純粋ロジック）
- `usecase`: `domain` のみに依存する。discordgo/pgx/kagome/net/httpなど外部技術には依存しない。必要な外部機能は自分でport interfaceを定義し、`infrastructure` がそれを実装する（依存性逆転）
- `infrastructure`: `usecase` が定義したport interfaceの実装を担う
- `interface`: 一番外側。discordgoのイベントやHTTPリクエストを受けて `usecase` を呼び出す

**新しい層をまたぐ依存を追加する前に、依存方向が `外側→内側` になっているか必ず確認すること。** `domain`/`usecase` が `infrastructure`/`interface` の型やパッケージをimportすることは絶対にない。

### 新機能追加の手順

Bot機能を1つ追加する際は以下の順序で実装する（詳細は [adr/0001_initial_architecture.md](./adr/0001_initial_architecture.md)）。

1. `internal/domain/<feature>/` に純粋ロジックを書く（あれば）
2. `internal/usecase/<feature>/` にユースケースとport interfaceを書く
3. 必要なら `internal/infrastructure/` にport実装を追加（既存実装を再利用できることが多い）
4. `internal/interface/discordbot/<feature>_handler.go` に `MessageHandler` 等を実装
5. `cmd/bot/main.go` の `router.Register...(...)` に1行追加

### ADR運用ルール

アーキテクチャ上の重要な決定（新しい依存の導入、レイヤー構成の変更、既存決定の撤回など）を行った際は、必ず `adr/NNNN_<slug>.md` として `adr/` ディレクトリに記録すること（連番はゼロ埋め4桁）。既存の決定を覆す場合は元のADRを書き換えず、新しいADRとして「なぜ変更するか」を記録する。

## 変更時のドキュメント更新

コードに変更を加えた際は、必要に応じて `README.md` と本 `CLAUDE.md` も追従して更新すること。特に以下は変更時に見直す。

- ディレクトリ構成やアーキテクチャに変更があった場合 → 両方のアーキテクチャ図・説明を更新
- 環境変数の追加/削除があった場合 → README.mdのセットアップ手順と `.env.example` を更新
- 新しいBot機能を追加した場合 → README.mdの機能一覧表を更新

## テスト方針

- 開発の基本はレッドグリーンテストによるTDD
- 作成した関数にはテストを入れ、テーブル駆動で書く
- エラーがある関数のテストケース名は「正常系: 」「異常系: 」で始め、異常系は「異常系: xxを入れるとyyなのでzzエラーになる」のように文章で書く
- `domain`/`usecase` 層のテストは外部依存（DB, Discord API, kagome辞書）なしで完結させる。`usecase` 層はport interfaceに対するフェイク実装（インメモリ）でテストする
- `infrastructure` 層のうち実データ検証が必要なもの（`postgres`, `morph`）は実際のライブラリ・DBを使ったテストを書いてよい

## Lint・テストコマンド

`make check`（`lint` → `vet` → `test`）でCI相当のチェックを一括実行できる。個々のコマンドは `Makefile` を参照。golangci-lintは `go.mod` の `tool` ディレクティブで管理しており、`go tool golangci-lint` で実行できる。設定は `.golangci.yml`。

コードを変更したら、コミット前に `make check` を通すこと。

## CI/CD

`.github/workflows/` にGitHub Actions定義がある（設計判断は [adr/0006_ci_build_pipeline.md](./adr/0006_ci_build_pipeline.md) 参照）。

- `100_test.yml` / `101_lint.yml`: PRごとに `go test` / `go vet` / `gofmt` / golangci-lintを実行する。golangci-lintは `golangci-lint-action` ではなく `go tool golangci-lint run` を使う（`go.mod` の `tool` ディレクティブで管理しているバージョンとCIのバージョンを一致させるため）。**ローカルの `make lint` が通るのにCIだけ落ちる場合、まずはこのコマンドの差異ではなくローカルの `go.sum`/tool バージョンがコミット済みのものと一致しているか疑うこと。**
- `900_build.yml`: mainブランチへのpush時に `linux/amd64,linux/arm64` のマルチアーキテクチャDockerイメージをビルドし `ghcr.io/usuyuki/usuyukis-discord-bot` へpushする。ビルド結果は `secrets.NOTIFY_DISCORD_WEBHOOK` 宛にDiscord通知される。マルチアーキビルドは `Dockerfile` 側のクロスコンパイル方式（QEMUエミュレーション回避、詳細は [adr/0009_docker_build_cross_compile.md](./adr/0009_docker_build_cross_compile.md)）で高速化している。
- `internal/infrastructure/postgres` に実DBを使うテストを追加した場合は、`100_test.yml` にPostgreSQLサービスコンテナを追加すること（現状はテストファイルがなくDB不要のため未設定）。

## コメント方針

- 作成した関数にはコメントを入れて何をしているかを書く
- 関数以外でも難しい処理やわかりにくい処理にはコメントを追加する
- コメントは「なぜそうなっているか」が非自明な場合に書く。読めば分かることは書かない
