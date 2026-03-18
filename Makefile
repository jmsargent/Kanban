.PHONY: validate

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
