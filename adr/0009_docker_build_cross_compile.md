# 0009. Dockerビルドをクロスコンパイル方式にしてマルチアーキビルドを高速化

## ステータス

決定（Accepted）

## コンテキスト

`900_build.yml` は [0006](./0006_ci_build_pipeline.md) の決定通り `linux/amd64,linux/arm64` のマルチアーキテクチャイメージをbuildxでbuildしていたが、実行時間が長かった。

原因はDockerfileの `FROM golang:1.26-bookworm`（プラットフォーム指定なし）で、buildxがターゲットプラットフォームごとにネイティブ実行しようとする構成になっていたこと。GitHub-hosted runner（amd64）上でarm64向けにビルドする際、Goコンパイラ自体をQEMUエミュレーション下で動かすことになり、ネイティブ実行に比べて大幅に遅くなっていた。umi.mikanのbackend/scheduler/subscriberは `--platform=$BUILDPLATFORM` + `GOARCH=$TARGETARCH` によるクロスコンパイル方式ですでにこの問題を回避している。

## 決定

`Dockerfile` のbuildステージを次のように変更する。

- `FROM golang:1.26-bookworm` → `FROM --platform=$BUILDPLATFORM golang:1.26-bookworm`\
  ビルドステージ自体は常にbuildホストのネイティブアーキテクチャ（amd64）で実行する。Goはクロスコンパイルに標準対応しているため、これでQEMUエミュレーションが不要になる。
- `ARG TARGETARCH` を追加し、`go build` に `GOOS=linux GOARCH=$TARGETARCH` を指定する。`TARGETARCH` はbuildxが自動的に埋める組み込み引数。
- `go mod download` と `go build` に `--mount=type=cache,target=/go/pkg/mod` / `--mount=type=cache,target=/root/.cache/go-build` を追加し、BuildKitのキャッシュマウントでGoモジュール・ビルドキャッシュをレイヤーキャッシュとは別に永続化する（umi.mikanのbackend Dockerfileと同じ手法）。

最終ステージ（`debian:bookworm-slim`）はバイナリの`COPY`のみのため、プラットフォーム指定は変更しない（ターゲットプラットフォームでネイティブに実行される）。

CGO依存のパッケージは使用していないため、`CGO_ENABLED=0`のクロスコンパイルに支障はない。

## 影響

- `900_build.yml` のマルチアーキテクチャビルドがQEMUエミュレーションなしで完了するようになり、実行時間が短縮される
- `docker/build-push-action` の `cache-from`/`cache-to`（`type=gha`, `type=registry`）による既存のレイヤーキャッシュに加え、BuildKitキャッシュマウントによりGoの依存解決・コンパイル結果もキャッシュされる
- ワークフローYAML自体の変更は不要（`Dockerfile` の変更のみ）
