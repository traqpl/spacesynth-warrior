.PHONY: all build wasm wasm-exec test ci docker-build deploy-agent run clean

BINARY       = spacesynth-warrior
WASM_OUT     = web/game.wasm
WASM_EXEC    = web/wasm_exec.js
PORT         ?= 8074
DOCKER_IMAGE ?= spacesynth-warrior:latest

GOROOT_WASM_EXEC = $(shell go env GOROOT)/lib/wasm/wasm_exec.js

ifeq ($(shell uname -s),Darwin)
BROWSER_OPEN = open
else
BROWSER_OPEN = xdg-open
endif

all: wasm run

wasm: wasm-exec
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o $(WASM_OUT) ./cmd/game/

wasm-exec:
	cp "$(GOROOT_WASM_EXEC)" "$(WASM_EXEC)"

build: wasm
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/server/

test:
	go test ./...

ci: test wasm build

docker-build:
	docker build -t $(DOCKER_IMAGE) .

deploy-agent: wasm
	DEPLOY_DIR="$${DEPLOY_DIR:-$$HOME/spacesynth-warrior}"; \
	mkdir -p "$$DEPLOY_DIR" && \
	docker build -t $(DOCKER_IMAGE) . && \
	cp compose.yaml "$$DEPLOY_DIR/compose.yaml" && \
	IMAGE_NAME=$(DOCKER_IMAGE) PORT=$(PORT) docker compose -f "$$DEPLOY_DIR/compose.yaml" up -d --force-recreate --remove-orphans

run: wasm
	@echo "Opening browser..."
	@$(BROWSER_OPEN) http://localhost:$(PORT) >/dev/null 2>&1 &
	go run ./cmd/server

clean:
	rm -f $(WASM_OUT) $(WASM_EXEC) $(BINARY)
