# syntax=docker/dockerfile:1.7

FROM --platform=linux/amd64 golang:1.26 AS build

WORKDIR /src

ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod
ENV GOTMPDIR=/root/.cache/go-tmp

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/go-tmp \
    mkdir -p "$GOCACHE" "$GOTMPDIR" && go mod download

COPY . .

# Build WASM game
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/go-tmp \
    mkdir -p "$GOCACHE" "$GOTMPDIR" && \
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js && \
    GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o web/game.wasm ./cmd/game/

# Build server binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/go-tmp \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server/

FROM scratch

WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=build /src/web /app/web

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]
