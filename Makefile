.PHONY: build run clean

WASM_OUT = web/game.wasm
WASM_EXEC = web/wasm_exec.js
GOROOT_WASM_EXEC = $(shell go env GOROOT)/lib/wasm/wasm_exec.js

build:
	cp "$(GOROOT_WASM_EXEC)" "$(WASM_EXEC)"
	GOOS=js GOARCH=wasm go build -o $(WASM_OUT) ./cmd/game

run: build
	@echo "Opening browser..."
	@open -a "Google Chrome" http://localhost:8080 &
	go run ./cmd/server

clean:
	rm -f $(WASM_OUT) $(WASM_EXEC)
