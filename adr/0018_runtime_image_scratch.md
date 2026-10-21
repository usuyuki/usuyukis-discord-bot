# 0018. 実行ステージのベースイメージを scratch にする

## ステータス

決定（Accepted）

## コンテキスト

`Dockerfile` の実行ステージは `debian:bookworm-slim` に `ca-certificates` を `apt-get install` する構成だった（[0009](./0009_docker_build_cross_compile.md) 時点）。

一方でバイナリは `CGO_ENABLED=0` でビルドした静的バイナリであり、実行時にOSのライブラリやシェルを必要としない。Debianレイヤーを含めるとイメージサイズが増え、Bot自体が使わないパッケージの脆弱性スキャン対象・パッチ追従対象も増える。umi.mikanのbackend/scheduler/subscriberはすでに `FROM scratch` で運用できている。

## 決定

実行ステージを `FROM scratch` に変更する。

| 必要なもの | 対応 |
|---|---|
| CA証明書（Discord API等へのHTTPS通信） | buildステージ（`golang:*-alpine`、ca-certificates同梱）の `/etc/ssl/certs/ca-certificates.crt` を `COPY` する |
| tzdata | buildステージで `apk add tzdata` し、`/usr/share/zoneinfo` を `COPY` する。JSTは `internal/domain/jst` で `time.FixedZone` により定義しており現状は参照しないが、依存ライブラリや将来の `time.LoadLocation` 利用に備えて入れておく |
| `migrations/` | 従来通り `/app/migrations` に `COPY` する（`cmd/bot/main.go` がカレントディレクトリ相対で参照するため `WORKDIR /app` を維持） |
| 実行ユーザー | scratchには `/etc/passwd` が無いため `USER 65534:65534`（nobody）を数値で指定し、非rootで実行する |

buildステージは `golang:*-bookworm` から `golang:*-alpine` に変更する（Goコンパイラが必要なためscratchにはできない）。`CGO_ENABLED=0` の静的バイナリなのでmusl/glibcの違いは影響せず、buildイメージが小さくなりCIでのpullが速くなる。umi.mikanのDockerfileとも揃う。

## 影響

- イメージサイズが小さくなり、OSパッケージ由来の脆弱性が実行イメージから消える
- コンテナ内にシェルが無くなるため `docker exec ... sh` での調査はできなくなる。調査が必要な場合はbuildステージをターゲットにしたイメージを別途ビルドする
- 将来CGO依存のパッケージを導入する場合は静的リンクできるか確認すること（動的リンクだとscratchでは起動しない）
