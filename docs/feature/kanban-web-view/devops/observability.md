# Observability: kanban-web-view

**Date**: 2026-03-29
**Status**: Draft
**Wave**: DEVOPS

---

## 1. Design Constraints

This is a single e2-micro VM with zero budget. No Prometheus, no Grafana Cloud, no Datadog. Observability uses only what ships with the OS and the Go standard library: structured logging to journald, a health check endpoint, and a simple cron-based monitor.

---

## 2. Structured Logging

### Log Format

All `kanban-web` log output uses Go's `log/slog` (stdlib, available since Go 1.21) in JSON format to stdout. systemd captures stdout to journald.

```go
slog.Info("request handled",
    "method", r.Method,
    "path", r.URL.Path,
    "status", statusCode,
    "duration_ms", elapsed.Milliseconds(),
    "remote_addr", r.RemoteAddr,
)
```

### Log Levels

| Level | When | Examples |
|-------|------|---------|
| ERROR | Unrecoverable failures | Git push failed, template render error, health check dependency down |
| WARN | Recoverable anomalies | Git pull failed (will retry next interval), auth token validation failed |
| INFO | Normal operations | Request handled, sync completed, deployment health check passed |
| DEBUG | Development diagnostics | Template data, config values, git command output |

Production runs at INFO level. Set `KANBAN_WEB_LOG_LEVEL=debug` for troubleshooting.

### What to Log

| Event | Level | Fields |
|-------|-------|--------|
| HTTP request | INFO | method, path, status, duration_ms, remote_addr |
| Git sync (pull) | INFO | result (ok/error), duration_ms |
| Git sync error | WARN | error message, will_retry_in |
| Task created | INFO | task_id, author_display_name (never log PAT) |
| Git push | INFO | task_id, result (ok/error), duration_ms |
| Git push error | ERROR | task_id, error message (never log PAT) |
| Server start | INFO | addr, repo_url, sync_interval, version |
| Server shutdown | INFO | reason (signal, error) |
| Auth token validation | INFO | result (ok/invalid), github_username (on success) |
| Auth token invalid | WARN | remote_addr (never log the token itself) |

### What NOT to Log

- GitHub PATs (never, under any circumstances)
- Cookie encryption keys
- Full request bodies
- User email addresses

### Querying Logs

```sh
# View blue instance logs
journalctl -u kanban-web-blue -f

# View last 100 lines
journalctl -u kanban-web-blue -n 100 --no-pager

# Filter errors
journalctl -u kanban-web-blue --priority=err

# Search for specific request
journalctl -u kanban-web-blue --grep="task_id.*abc123"

# Logs since last deployment
journalctl -u kanban-web-blue --since="2026-03-29 14:00:00"
```

---

## 3. Health Check Endpoint

### `GET /healthz`

Returns 200 if the server is operational and can read from the git clone.

```json
{
  "status": "ok",
  "version": "0.1.0",
  "uptime_seconds": 3600,
  "last_sync": "2026-03-29T14:30:00Z",
  "last_sync_ok": true
}
```

Returns 503 if the server cannot read task files from the local clone:

```json
{
  "status": "degraded",
  "version": "0.1.0",
  "error": "git clone not readable"
}
```

### Implementation Notes

- The health check reads a minimal file from the git clone (e.g., checks that `.kanban/` directory exists)
- It does NOT perform a `git pull` -- that would add latency and could cause rate limiting
- The `last_sync` and `last_sync_ok` fields report the result of the most recent background sync
- nginx health checks use this endpoint during blue/green deployment

---

## 4. Monitoring (Zero Budget)

### Option A: Cron-Based Health Monitor

A simple shell script runs every 5 minutes via cron. If the health check fails 3 consecutive times, it restarts the service and logs a message.

#### `/var/kanban/bin/health-monitor.sh`

```sh
#!/bin/sh
# Simple health monitor for kanban-web.
# Checks /healthz, restarts service on 3 consecutive failures.
# Install: crontab -e -> */5 * * * * /var/kanban/bin/health-monitor.sh

FAIL_FILE="/tmp/kanban-web-health-fails"
MAX_FAILS=3

# Determine active instance
ACTIVE_PORT=$(grep -v '#' /etc/nginx/sites-available/kanban-web | grep 'server 127.0.0.1' | grep -oP ':\K[0-9]+' | head -1)
SERVICE="kanban-web-$([ "$ACTIVE_PORT" = "8080" ] && echo "blue" || echo "green")"

if curl -sf --max-time 5 "http://127.0.0.1:${ACTIVE_PORT}/healthz" > /dev/null 2>&1; then
  # Success -- reset counter
  rm -f "$FAIL_FILE"
else
  # Failure -- increment counter
  FAILS=$(cat "$FAIL_FILE" 2>/dev/null || echo 0)
  FAILS=$((FAILS + 1))
  echo "$FAILS" > "$FAIL_FILE"

  if [ "$FAILS" -ge "$MAX_FAILS" ]; then
    logger -t kanban-monitor "ALERT: ${SERVICE} failed ${MAX_FAILS} consecutive health checks. Restarting."
    sudo systemctl restart "$SERVICE"
    rm -f "$FAIL_FILE"
  else
    logger -t kanban-monitor "WARN: ${SERVICE} health check failed (${FAILS}/${MAX_FAILS})"
  fi
fi
```

```sh
# Install
chmod +x /var/kanban/bin/health-monitor.sh
crontab -l | { cat; echo "*/5 * * * * /var/kanban/bin/health-monitor.sh"; } | crontab -
```

### Option B: UptimeRobot (Free Tier)

[UptimeRobot](https://uptimerobot.com/) offers 50 free monitors with 5-minute intervals. Configure a single HTTP monitor:

| Setting | Value |
|---------|-------|
| URL | `https://YOUR_IP_OR_DOMAIN/healthz` |
| Interval | 5 minutes |
| Alert contacts | Email (free) |
| Keyword | `"status":"ok"` |

This provides external monitoring (detects if the entire VM is down) and email alerts at zero cost. It complements the cron-based internal monitor.

**Recommendation**: Use both. Cron handles auto-restart. UptimeRobot handles "VM is completely down" alerting.

---

## 5. SLO Definition (Informal)

For a student project, formal SLOs with error budgets are unnecessary. Instead, define practical availability targets:

| Metric | Target | Measurement |
|--------|--------|-------------|
| Availability | 95% (36.5 hours downtime/month acceptable) | UptimeRobot uptime report |
| Response time (board view) | < 2 seconds p95 | Health check logs duration_ms |
| Sync freshness | < 2 minutes behind remote | last_sync field in /healthz |

These are aspirational, not contractual. No automated error budget enforcement. Review UptimeRobot monthly report to track trends.

---

## 6. nginx Access Logs

nginx default access log at `/var/log/nginx/access.log` provides request-level data. No additional configuration needed.

Useful queries:

```sh
# Top 10 most requested paths
awk '{print $7}' /var/log/nginx/access.log | sort | uniq -c | sort -rn | head -10

# 5xx errors in last hour
awk '$9 ~ /^5/ && $4 > "[29/Mar/2026:13"' /var/log/nginx/access.log

# Request rate per minute
awk '{print $4}' /var/log/nginx/access.log | cut -d: -f1-3 | uniq -c | tail -20
```

---

## 7. Disk and Resource Monitoring

The e2-micro has limited resources. Monitor them with a weekly cron job:

```sh
# Add to crontab: run weekly
0 9 * * 1 /var/kanban/bin/resource-check.sh
```

#### `/var/kanban/bin/resource-check.sh`

```sh
#!/bin/sh
# Weekly resource check. Logs warnings if approaching limits.

DISK_USAGE=$(df /var/kanban --output=pcent | tail -1 | tr -d ' %')
if [ "$DISK_USAGE" -gt 80 ]; then
  logger -t kanban-resources "WARN: Disk usage at ${DISK_USAGE}%"
fi

# Git clone size
CLONE_SIZE=$(du -sh /var/kanban/repo | awk '{print $1}')
logger -t kanban-resources "INFO: Git clone size: ${CLONE_SIZE}, Disk usage: ${DISK_USAGE}%"
```

---

## 8. Incident Response (Lean)

For a single-person student project, formal incident response is overhead. Instead:

| Symptom | Check | Fix |
|---------|-------|-----|
| Site unreachable | `curl https://IP/healthz` | Check VM status in GCP console; `sudo systemctl status nginx` |
| 502 Bad Gateway | `sudo systemctl status kanban-web-blue` | Restart service; check logs with `journalctl` |
| Board shows stale data | Check `last_sync` in `/healthz` response | SSH in, run `sudo -u kanban git -C /var/kanban/repo pull` manually |
| Git push fails for users | Check `journalctl -u kanban-web-blue --grep="push"` | Likely PAT permissions; user needs repo write access |
| Disk full | `df -h` | Clear old logs: `sudo journalctl --vacuum-time=7d` |
| High memory | `free -h` and `top` | Restart service; investigate if the git clone is too large |

---

## 9. What This Observability Setup Does NOT Include

Intentionally excluded to match zero-budget, single-VM constraints:

- **No metrics collection** (no Prometheus/StatsD) -- use log analysis instead
- **No distributed tracing** -- single service, no need
- **No dashboards** -- UptimeRobot provides the only visual
- **No APM** -- structured logs plus health check are sufficient for this scale
- **No log aggregation** -- journald on the VM is the log store

If the project grows beyond the e2-micro, revisit with: GCP Cloud Logging (free tier: 50 GB/month), GCP Cloud Monitoring (free tier: basic metrics), or Grafana Cloud free tier (10K metrics, 50 GB logs).
