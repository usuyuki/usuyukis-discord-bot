# 0006. GitHub ActionsによるCI/CDパイプラインの整備

## ステータス

決定（Accepted）

## コンテキスト

これまで本リポジトリには `.github/workflows` が存在せず、`make check`（lint → vet → test）はローカル実行のみで、PRマージ前の自動検証もコンテナイメージの自動publishもなかった。姉妹プロジェクトである [umi.mikan](https://github.com/project-mikan/umi.mikan) がGo backendに対して行っているCI構成（PRごとのGo Test/Lint、mainブランチpush時のマルチアーキテクチャDockerイメージbuild & push、PR自動ラベル付与、Dependabot）を参考に、同等の仕組みを本リポジトリにも導入する。

## 決定

### 1. ワークフローファイルの命名規則をumi.mikanに合わせる

`NNN_<用途>.yml` の3桁連番プレフィックスで実行順・カテゴリを可視化する（umi.mikanは `100_backend_test.yml` のようにサービス名を含むが、本リポジトリは単一Goモジュールのため用途名のみとする）。

- `000_labeler.yml`: PR自動ラベル付与
- `100_test.yml`: `go test`（カバレッジ計測 + Codecovアップロード含む）
- `101_lint.yml`: `go vet` / `gofmt` / golangci-lint
- `900_build.yml`: mainブランチpush時のDockerイメージbuild & push（GHCR）

### 2. golangci-lintはActionではなく `go tool golangci-lint` で実行する

`golangci-lint-action` を使うとActionが指定するバージョンとローカルの `go.mod` tool ディレクティブが管理するバージョンがずれ得る（umi.mikanのiOS CIで `--strict` オプションの有無によりローカルとCIの結果が食い違った教訓と同種の問題）。本リポジトリはすでに `go.mod` の `tool` ディレクティブでgolangci-lint v2を固定管理し、`Makefile` の `lint`/`format` ターゲットも `go tool golangci-lint` を使っているため、CIも同じコマンドを使い「ローカルで通ったのにCIだけ落ちる／その逆」を構造的に防ぐ。

### 3. テストジョブにPostgreSQLサービスコンテナは含めない

`internal/infrastructure/postgres` には現時点でテストファイルが存在せず（`domain`/`usecase` はport interfaceのフェイク実装でテストする方針のため）、`go test ./...` はDBなしで完結する。将来 `postgres`/`morph` 層に実DBを使う統合テストを追加した場合は、このワークフローにPostgreSQLサービスコンテナを追加すること。

### 4. Dockerイメージは `linux/amd64,linux/arm64` のマルチアーキテクチャでGHCRへpush

自宅サーバーでの運用（Raspberry Pi等のarm64環境を含む可能性）を考慮し、umi.mikanのbackend/scheduler/subscriberと同様にマルチアーキテクチャbuildとする。イメージ名は `ghcr.io/usuyuki/usuyukis-discord-bot`。GitHub Actionsのデフォルト `GITHUB_TOKEN` でGHCRへログインするため、`900_build.yml` のjobに明示的に `permissions: packages: write` を付与する（umi.mikan側はリポジトリ設定のデフォルト権限に依存しているが、本リポジトリでは設定に依存せずワークフロー側で明示する）。

Discord通知（`sarisia/actions-status-discord`）はumi.mikan側では専用のwebhook secretを使っているが、本リポジトリにはその secret が未設定のため今回は導入しない。必要になれば別途 `NOTIFY_DISCORD_WEBHOOK` 等のsecretを追加した上でADRを更新せず本ADRの範囲内の追加として導入してよい。

### 5. Dependabotで `gomod` / `docker` / `github-actions` を月次更新

umi.mikanの `dependabot.yml` に倣い、3つのpackage-ecosystemを月次(Asia/Tokyo)でグループ更新する。

## 影響

- PRを出すと `go test` / `go vet` / `gofmt` / golangci-lint が自動実行され、レビュー前に機械的な問題を検出できる
- mainへのマージ後、GHCRに `ghcr.io/usuyuki/usuyukis-discord-bot:latest` 等のイメージが自動publishされる。本番運用でこのイメージをpullする構成に変更する場合は、別途デプロイ手順のドキュメント化・ADR化を検討すること
- CIが `go.mod` の `tool` ディレクティブに追従するため、golangci-lintのバージョンアップは通常のGo依存更新（Dependabotの `gomod` グループ、または手動の `go get -tool`）で完結し、ワークフローYAMLの変更は不要
