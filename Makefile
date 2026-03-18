.PHONY: validate acceptance ci release-snapshot release tag-dry help

## validate: run all quality gates locally (mirrors CI validate-and-build job)
validate:
	@echo "[0/5] check-versions"
	@cicd/check-versions.sh
	@echo "[1/5] go test ./internal/..."
	@go test ./internal/...
	@echo "[2/5] golangci-lint run"
	@golangci-lint run
	@echo "[3/5] go-arch-lint check"
	@go-arch-lint check
	@echo "[4/5] go build ./..."
	@go build ./...
	@echo "PASS"

## acceptance: build kanban binary and run acceptance tests
acceptance:
	@echo "Building kanban binary..."
	@go build -o kanban ./cmd/kanban
	@echo "Running acceptance tests..."
	@KANBAN_BIN=./kanban go test ./tests/acceptance/...

## ci: run validate then acceptance (mirrors CI pipeline locally)
ci:
	@make validate && make acceptance

## release-snapshot: build all cross-compile targets locally without publishing
release-snapshot:
	@goreleaser release --snapshot --clean --config cicd/goreleaser.yml

## release: publish a release via goreleaser (requires GITHUB_TOKEN)
release:
	@goreleaser release --clean --config cicd/goreleaser.yml

## tag-dry: dry-run semantic version tagging
tag-dry:
	@go-semver-release release --dry-run

## help: list available make targets
help:
	@echo "Available targets:"
	@echo "  validate        run all quality gates (mirrors CI validate-and-build)"
	@echo "  acceptance      build binary and run acceptance tests"
	@echo "  ci              run validate then acceptance"
	@echo "  release-snapshot build cross-compile targets without publishing"
	@echo "  release         publish a release via goreleaser"
	@echo "  tag-dry         dry-run semantic version tagging"
	@echo "  help            list available make targets"
