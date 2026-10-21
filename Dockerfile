FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETARCH
# golang:alpineにはtzdataが同梱されていない
RUN apk add --no-cache tzdata
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o /out/bot ./cmd/bot

# scratchで動くのはCGO_ENABLED=0の静的バイナリのため。tzdataは現状未使用（JSTはFixedZone）だが保険で入れる
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
WORKDIR /app
COPY --from=build /out/bot ./bot
# 非rootで読めるようにする。444だとディレクトリに実行ビットが付かず中を読めない
COPY --from=build --chmod=a+rX /src/migrations ./migrations
# 65534 = nobody（scratchに/etc/passwdが無いため数値指定）
USER 65534:65534
ENTRYPOINT ["/app/bot"]
