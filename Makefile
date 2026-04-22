PORT ?= 8090
BIND ?= 0.0.0.0
MODEL ?= github-copilot/claude-opus-4.6
WORKSPACE ?= /workspace
RUN_DIR ?= .gi-run
BIN_DIR ?= bin
BIN ?= $(BIN_DIR)/gi
DB ?= $(RUN_DIR)/gi.db
LOG ?= $(RUN_DIR)/gi.log
PID ?= $(RUN_DIR)/gi.pid
LISTEN ?=

.PHONY: build-web build test vet start stop restart status logs run clean

build-web:
	bun run build:web

build: build-web
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd/gi

test:
	go test ./...

vet:
	go vet ./...

bun-checks:
	bun run check:hook-tdz

run: build
	$(BIN) $(if $(LISTEN),-listen $(LISTEN),-bind $(BIND) -port $(PORT)) -model $(MODEL) -db $(DB) -workspace $(WORKSPACE)

start: build
	mkdir -p $(RUN_DIR)
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		echo "Gi already running with PID $$(cat $(PID))"; \
		exit 0; \
	fi
	$(abspath $(BIN)) $(if $(LISTEN),-listen $(LISTEN),-bind $(BIND) -port $(PORT)) \
		-model $(MODEL) \
		-db $(abspath $(DB)) \
		-workspace $(WORKSPACE) \
		-log-file $(abspath $(LOG)) \
		-pid-file $(abspath $(PID)) \
		>/dev/null 2>&1 </dev/null &
	@sleep 2
	@$(MAKE) --no-print-directory status BIND=$(BIND) PORT=$(PORT)

stop:
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		kill $$(cat $(PID)) && echo "Stopped Gi ($$(cat $(PID)))"; \
		rm -f $(PID); \
	else \
		echo "Gi is not running"; \
	fi

restart: stop start

status:
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		echo "Gi running on $(if $(LISTEN),$(LISTEN),$(BIND):$(PORT)) with PID $$(cat $(PID))"; \
		ss -ltnp | grep ':$(PORT)' || true; \
	else \
		echo "Gi is not running"; \
		exit 1; \
	fi

logs:
	@mkdir -p $(RUN_DIR)
	tail -f $(LOG)

clean:
	rm -rf $(RUN_DIR) $(BIN_DIR)
