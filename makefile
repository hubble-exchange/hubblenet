.PHONY: run
run:
	./scripts/run_local.sh

.PHONY: upgrade
upgrade:
	./scripts/upgrade_local.sh

.PHONY: logs
logs:
	./scripts/show_logs.sh

.PHONY: logs-1
logs-1:
	./scripts/show_logs.sh 1

.PHONY: test
test:
	go test ./plugin/...
	go test ./precompile/...
	go test ./hubbleutils/...
	go test ./orderbook/...

.PHONY: tidy
tidy:
	go mod download && go mod tidy -compat=1.21.7

.PHONY: fmt
fmt:
	go fmt ./...