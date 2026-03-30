# CI/CD Pipeline: kanban-web-view

**Date**: 2026-03-29
**Status**: Draft
**Wave**: DEVOPS

---

## 1. Pipeline Strategy

The existing CircleCI pipeline at `cicd/config.yml` builds and releases the `kanban` CLI binary. The `kanban-web` binary extends this pipeline -- both binaries are built, tested, and released from the same CI run.

**Branching model**: Trunk-based development on `main` (unchanged). Every push to `main` triggers the full pipeline.

---

## 2. Extended CircleCI Pipeline

### Updated Workflow

```
validate-and-build  -->  acceptance  -->  tag-and-release  -->  smoke-test (matrix)
                                                            -->  deploy-web (NEW)
```

The `deploy-web` job runs after `tag-and-release` succeeds. It cross-compiles the `kanban-web` binary for `linux/amd64`, uploads it to the GCP VM via `scp`, and executes the blue/green deployment script.

### Changes to Existing Jobs

#### validate-and-build (modified)

Add build step for kanban-web binary:

```yaml
- run:
    name: Build (kanban-cli)
    command: make ci-build-cli
- run:
    name: Build (kanban-web)
    command: make ci-build-web
- persist_to_workspace:
    root: .
    paths:
      - kanban
      - kanban-web
```

#### acceptance (modified)

Web-specific acceptance tests run alongside existing CLI acceptance tests:

```yaml
- run:
    name: E2E Tests (CLI)
    command: make ci-e2e-tests
- run:
    name: E2E Tests (Web)
    command: make ci-e2e-tests-web
```

### New Jobs

#### deploy-web (new)

```yaml
deploy-web:
  executor: go-executor
  steps:
    - checkout
    - attach_workspace:
        at: .
    - run:
        name: Cross-compile kanban-web for linux/amd64
        command: GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kanban-web-linux -ldflags="-s -w" ./cmd/kanban-web
    - run:
        name: Deploy to GCP (blue/green)
        command: cicd/deploy-web.sh
```

### Updated Workflow Section

```yaml
workflows:
  ci-cd:
    jobs:
      - validate-and-build:
          filters:
            branches:
              only: main
      - acceptance:
          requires:
            - validate-and-build
      - tag-and-release:
          requires:
            - acceptance
          context:
            - release-context
      - smoke-test:
          requires:
            - tag-and-release
          context:
            - release-context
          matrix:
            parameters:
              channel:
                - binary
                - go-install
      - deploy-web:
          requires:
            - tag-and-release
          context:
            - deploy-context
          filters:
            branches:
              only: main
```

---

## 3. New Makefile Targets

```makefile
## ci-build-cli: build the kanban CLI binary (CI step)
ci-build-cli:
	@go build -o kanban ./cmd/kanban-cli

## ci-build-web: build the kanban-web binary (CI step)
ci-build-web:
	@go build -o kanban-web ./cmd/kanban-web

## ci-e2e-tests-web: run web acceptance tests via gotestsum (CI step)
ci-e2e-tests-web:
	@KANBAN_WEB_BIN="$(CURDIR)/kanban-web" gotestsum --format testname -- ./tests/acceptance/web/...

## ci-build (backward compat): build both binaries
ci-build: ci-build-cli ci-build-web
```

---

## 4. goreleaser Extension

Add a second build target to `cicd/goreleaser.yml`:

```yaml
builds:
  - id: kanban
    main: ./cmd/kanban-cli
    binary: kanban
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

  - id: kanban-web
    main: ./cmd/kanban-web
    binary: kanban-web
    env:
      - CGO_ENABLED=0
    goos: [linux]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
```

**Note**: `kanban-web` is built only for Linux (the deployment target). No need for darwin/windows builds -- it is a server binary, not a CLI tool distributed to end users.

The Homebrew formula and `go install` instructions remain for the `kanban` CLI only. `kanban-web` is deployed directly, not distributed via package managers.

---

## 5. Deployment Script: `cicd/deploy-web.sh`

```sh
#!/bin/sh
# Blue/green deployment for kanban-web on GCP Compute Engine.
#
# Required environment variables (from CircleCI deploy-context):
#   GCP_SSH_KEY_BASE64  -- base64-encoded SSH private key for the VM
#   GCP_VM_IP           -- external IP of the kanban-web VM
#   GCP_VM_USER         -- SSH user on the VM (default: deploy)
#
# Usage: cicd/deploy-web.sh

set -euo pipefail

VM_USER="${GCP_VM_USER:-deploy}"
VM_IP="${GCP_VM_IP:?GCP_VM_IP is required}"
BINARY="kanban-web-linux"
REMOTE_DIR="/var/kanban/bin"

# --- Determine which instance is active ---
ACTIVE=$(ssh -o StrictHostKeyChecking=no "${VM_USER}@${VM_IP}" \
  "grep -v '#' /etc/nginx/sites-available/kanban-web | grep 'server 127.0.0.1' | grep -oP ':\K[0-9]+'")

if [ "$ACTIVE" = "8080" ]; then
  DEPLOY_TARGET="green"
  DEPLOY_PORT="8081"
  DEPLOY_SERVICE="kanban-web-green"
else
  DEPLOY_TARGET="blue"
  DEPLOY_PORT="8080"
  DEPLOY_SERVICE="kanban-web-blue"
fi

echo "Active instance on :${ACTIVE}. Deploying to ${DEPLOY_TARGET} (:${DEPLOY_PORT})."

# --- Upload binary ---
scp -o StrictHostKeyChecking=no "${BINARY}" "${VM_USER}@${VM_IP}:/tmp/kanban-web-new"

# --- Deploy to inactive instance ---
ssh -o StrictHostKeyChecking=no "${VM_USER}@${VM_IP}" << REMOTE
  set -euo pipefail

  # Move binary into place
  sudo mv /tmp/kanban-web-new ${REMOTE_DIR}/kanban-web-${DEPLOY_TARGET}
  sudo chmod 755 ${REMOTE_DIR}/kanban-web-${DEPLOY_TARGET}
  sudo chown kanban:kanban ${REMOTE_DIR}/kanban-web-${DEPLOY_TARGET}

  # Restart the target service
  sudo systemctl restart ${DEPLOY_SERVICE}
  sleep 2

  # Health check
  if ! curl -sf http://127.0.0.1:${DEPLOY_PORT}/healthz > /dev/null; then
    echo "FAIL: health check on ${DEPLOY_TARGET} (:${DEPLOY_PORT})"
    sudo journalctl -u ${DEPLOY_SERVICE} --no-pager -n 20
    exit 1
  fi

  echo "Health check passed on ${DEPLOY_TARGET}."

  # Switch nginx upstream
  sudo sed -i 's|^\(\s*\)server 127.0.0.1:${ACTIVE};|#\1server 127.0.0.1:${ACTIVE};|' /etc/nginx/sites-available/kanban-web
  sudo sed -i 's|^#\(\s*\)server 127.0.0.1:${DEPLOY_PORT};|\1server 127.0.0.1:${DEPLOY_PORT};|' /etc/nginx/sites-available/kanban-web
  sudo nginx -t && sudo nginx -s reload

  echo "Traffic switched to ${DEPLOY_TARGET} (:${DEPLOY_PORT})."

  # Stop old instance (saves memory on e2-micro)
  OLD_SERVICE="kanban-web-$([ "${DEPLOY_TARGET}" = "blue" ] && echo "green" || echo "blue")"
  sudo systemctl stop \${OLD_SERVICE}

  echo "Deployment complete. Active: ${DEPLOY_TARGET}."
REMOTE
```

---

## 6. CircleCI Context: deploy-context

A new CircleCI context `deploy-context` is required. It must contain:

| Variable | Description |
|----------|-------------|
| `GCP_SSH_KEY_BASE64` | Base64-encoded SSH private key for the deploy user on the VM |
| `GCP_VM_IP` | External IP address of the kanban-web VM |
| `GCP_VM_USER` | SSH user on the VM (default: `deploy`) |

### Deploy User Setup (on the VM)

```sh
# Create a deploy user with sudo rights for specific commands
sudo useradd -r -m -s /bin/bash deploy
sudo mkdir -p /home/deploy/.ssh
sudo chmod 700 /home/deploy/.ssh

# Add CI's public key to authorized_keys
echo "PUBLIC_KEY_HERE" | sudo tee /home/deploy/.ssh/authorized_keys
sudo chmod 600 /home/deploy/.ssh/authorized_keys
sudo chown -R deploy:deploy /home/deploy/.ssh

# Grant deploy user limited sudo (no password) for deployment commands only
echo 'deploy ALL=(ALL) NOPASSWD: /bin/mv, /bin/chmod, /bin/chown, /bin/systemctl, /usr/sbin/nginx, /bin/sed' \
  | sudo tee /etc/sudoers.d/deploy
```

---

## 7. Test Strategy

### Layer Mapping

| Layer | Scope | Runner | Gate |
|-------|-------|--------|------|
| Domain | Pure unit tests (existing) | `go test ./internal/domain/...` | Pre-commit + CI |
| Use cases | Mock port tests (existing + new AddTaskAndPush) | `go test ./internal/usecases/...` | Pre-commit + CI |
| Web adapter | `httptest` handler tests | `go test ./internal/adapters/web/...` | Pre-commit + CI |
| Git adapter | Integration tests with `t.TempDir()` + `git init` | `go test ./internal/adapters/git/...` | Pre-commit + CI |
| CLI E2E | Binary subprocess tests (existing) | `go test ./tests/acceptance/...` | CI |
| Web E2E | Start kanban-web, HTTP client tests against it | `go test ./tests/acceptance/web/...` | CI |

### Web E2E Tests

The web E2E tests start the `kanban-web` binary as a subprocess (same pattern as CLI E2E), initialize a test git repo in `t.TempDir()`, and make HTTP requests against it:

1. `GET /board` returns 200 with board HTML
2. `GET /task/{id}` returns 200 with task detail HTML
3. `GET /healthz` returns 200
4. `POST /task` without auth returns 401 or redirect to token entry
5. Full add-task flow (set auth cookie, POST /task, verify file created)

No browser automation (Selenium/Playwright) -- tests use Go's `net/http` client. This keeps tests fast and dependency-free.

---

## 8. Rollback Procedure

Rollback is the primary advantage of blue/green deployment. The procedure is:

### Automatic Rollback (in deploy script)

If the health check fails after deploying to the inactive instance, the script exits with error. No traffic switch occurs. The active instance continues serving. CircleCI reports the job as failed.

### Manual Rollback (post-traffic-switch)

If a problem is detected after the traffic switch:

```sh
# SSH to VM
gcloud compute ssh kanban-web --zone=us-central1-a

# Switch nginx back to the previous instance
# Edit /etc/nginx/sites-available/kanban-web: uncomment old, comment new
sudo nginx -t && sudo nginx -s reload

# Restart the old instance if it was stopped
sudo systemctl start kanban-web-blue  # or green
```

Time to rollback: under 30 seconds (one SSH command + nginx reload).

### Rollback Checklist

- [ ] Previous binary still present on VM (`/var/kanban/bin/kanban-web-{blue,green}`)
- [ ] Previous systemd service is enabled and can start
- [ ] nginx upstream config is easily toggled
- [ ] No database migrations to revert (stateless application)
- [ ] Git clone state is shared -- no rollback needed for data

---

## 9. Pipeline Quality Gates

| Stage | Gate | Type | Threshold |
|-------|------|------|-----------|
| Pre-commit (local) | Arch lint, static analysis, unit tests, build both binaries | Blocking | All pass |
| CI: validate-and-build | Arch check, lint, vet, vulncheck, unit tests, build both binaries | Blocking | All pass |
| CI: acceptance | CLI E2E tests, Web E2E tests | Blocking | All pass |
| CI: tag-and-release | Semantic version tag + goreleaser release | Blocking | Tag created |
| CI: deploy-web | Cross-compile, upload, health check, traffic switch | Blocking | Health check pass |
| Production | Health endpoint responds 200 | Advisory | Manual check |
