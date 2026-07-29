# 0008. 本番環境のdocker composeはGHCRのビルド済みイメージを使う

## ステータス

決定（Accepted）

## コンテキスト

[0006](./0006_ci_build_pipeline.md) で `main` ブランチへのpush時にマルチアーキテクチャ（`linux/amd64`/`linux/arm64`）Dockerイメージを `ghcr.io/usuyuki/usuyukis-discord-bot` へ自動publishする仕組みを導入したが、`compose.yml` の `bot` サービスは `build: .` のみを指定しており、`docker compose up -d --build` のたびに本番環境（自宅サーバー等）でもソースからのビルドが発生していた。0006のADR内でも「本番運用でこのイメージをpullする構成に変更する場合は、別途デプロイ手順のドキュメント化・ADR化を検討すること」と課題化されていた。

本番サーバー上でのビルドは以下の点で無駄が大きい。

- CIで既にビルド・マルチアーキ対応済みのイメージが存在するにもかかわらず、本番サーバーのCPU/メモリを使って重複してビルドする
- 本番サーバー上にGoツールチェイン相当のビルド環境を用意する必要がある（Dockerfileのbuildステージ自体はマルチステージビルドでイメージ内に閉じるが、ビルド自体の負荷はサーバーにかかる）
- ビルドとデプロイの手順が同じコマンド（`docker compose up -d --build`）に混在し、「今動いているイメージがどのコミットに対応するか」が本番サーバー側のGitチェックアウト状態に依存してしまう

## 決定

`compose.yml` の `bot` サービスに `image` と `build` を両方指定する。

```yaml
services:
  bot:
    image: ghcr.io/usuyuki/usuyukis-discord-bot:${BOT_IMAGE_TAG:-latest}
    build: .
```

- **本番運用**は `docker compose pull bot && docker compose up -d` を使う。`image` が指定されているため、ローカルビルドを一切行わずGHCRからマルチアーキイメージをpullして起動する
- **開発**は従来どおり `docker compose up -d --build` を使う。`build` も指定されているため、`--build` を付けた場合のみローカルの `Dockerfile` からビルドしたイメージで `image` に指定したタグを上書きする
- イメージタグは `.env` の `BOT_IMAGE_TAG`（デフォルト `latest` = mainブランチ最新ビルド）で切り替え可能にする。特定コミットのイメージを固定したい場合は `900_build.yml` がpushする `main-<sha>` タグを指定する
- Docker Compose標準の挙動（`image`と`build`両方指定時、`pull`/`up`（`--build`なし）はレジストリからpull、`build`/`up --build`はローカルビルドしてそのタグに上書き）にそのまま乗るだけで、専用のoverrideファイル（`compose.prod.yml`等）は用意しない

## 影響

- 本番サーバーでのデプロイ手順が「`git pull` してソースからビルド」から「`docker compose pull` してビルド済みイメージを取得」に変わる。デプロイ時にサーバー上にソースの最新チェックアウトを保持し続ける必要がなくなる（`compose.yml`と`.env`さえあれば起動可能）
- `docker compose up -d --build` を本番で誤って実行すると、そのサーバー上のソース状態でローカルビルドしたイメージが `latest` タグ相当を上書きしてしまう。README.mdに本番運用と開発運用のコマンドの違いを明記して周知する
- README.mdのセットアップ手順に本番運用（pull）と開発運用（build）を分けて追記した
