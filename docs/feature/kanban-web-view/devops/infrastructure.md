# Infrastructure: kanban-web-view

**Date**: 2026-03-29
**Status**: Draft
**Wave**: DEVOPS

---

## 1. VM Provisioning

### GCP Compute Engine Specification

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Machine type | e2-micro | Free tier (0.25 vCPU, 1 GB RAM) |
| Region | us-central1-a | Free tier eligible |
| OS image | Debian 12 (Bookworm) | Stable, minimal, nginx in default repos |
| Disk | 30 GB standard persistent | Free tier limit |
| External IP | Ephemeral static | Free while attached to running VM |

### Provisioning Steps (Manual -- No Terraform)

Terraform is rejected for this project. A single e2-micro VM with manual setup is simpler, faster, and appropriate for a student project with one server. Infrastructure-as-code adds value when there are multiple environments or reproducibility requirements -- neither applies here.

```sh
# 1. Create VM
gcloud compute instances create kanban-web \
  --zone=us-central1-a \
  --machine-type=e2-micro \
  --image-family=debian-12 \
  --image-project=debian-cloud \
  --boot-disk-size=30GB \
  --tags=http-server,https-server

# 2. Reserve the ephemeral IP as static (free while attached)
gcloud compute addresses create kanban-web-ip \
  --region=us-central1 \
  --addresses=$(gcloud compute instances describe kanban-web \
    --zone=us-central1-a --format='get(networkInterfaces[0].accessConfigs[0].natIP)')
```

### Initial Server Setup

```sh
# SSH into the VM
gcloud compute ssh kanban-web --zone=us-central1-a

# Update and install dependencies
sudo apt update && sudo apt upgrade -y
sudo apt install -y nginx git certbot python3-certbot-nginx

# Create application user (no login shell, owns the app directory)
sudo useradd -r -s /usr/sbin/nologin -m -d /var/kanban kanban

# Create directories
sudo mkdir -p /var/kanban/repo
sudo mkdir -p /var/kanban/bin
sudo chown -R kanban:kanban /var/kanban
```

---

## 2. SSH Deploy Key for Git Clone

The VM needs read access to the GitHub repo for cloning and pulling. For public repos this is not required, but it is needed for private repos and for push operations.

### Setup

```sh
# Generate deploy key (on the VM, as the kanban user)
sudo -u kanban ssh-keygen -t ed25519 -C "kanban-web-deploy" -f /var/kanban/.ssh/id_ed25519 -N ""

# Add to GitHub repo: Settings -> Deploy Keys -> Add deploy key
# Title: kanban-web-view (production)
# Key: contents of /var/kanban/.ssh/id_ed25519.pub
# Allow write access: NO (push uses user PAT, not deploy key)
```

For push operations, the user's GitHub PAT is injected into the HTTPS remote URL per-push (as designed in the architecture). The deploy key is read-only.

**For public repos**: The deploy key is optional. Clone and pull work over HTTPS without authentication. Only configure the deploy key if the repository is private.

---

## 3. Firewall Rules

```sh
# Allow HTTP (for Let's Encrypt challenges and redirect)
gcloud compute firewall-rules create allow-http \
  --allow=tcp:80 \
  --target-tags=http-server \
  --source-ranges=0.0.0.0/0

# Allow HTTPS
gcloud compute firewall-rules create allow-https \
  --allow=tcp:443 \
  --target-tags=https-server \
  --source-ranges=0.0.0.0/0

# SSH is allowed by default on GCP (port 22)
# Ports 8080/8081 are NOT exposed -- nginx proxies internally
```

---

## 4. nginx Configuration

### Main Config: `/etc/nginx/sites-available/kanban-web`

```nginx
# Blue/green upstream -- toggle active line during deployment
upstream kanban_backend {
    server 127.0.0.1:8080;  # blue (active)
    # server 127.0.0.1:8081;  # green (standby)
}

server {
    listen 80;
    server_name _;

    # Let's Encrypt challenge path
    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    # Redirect all other HTTP to HTTPS
    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name _;

    # TLS certificates (managed by certbot)
    ssl_certificate /etc/letsencrypt/live/DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/DOMAIN/privkey.pem;

    # TLS hardening
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:10m;

    # Proxy to kanban-web
    location / {
        proxy_pass http://kanban_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts suited to e2-micro performance
        proxy_connect_timeout 5s;
        proxy_read_timeout 30s;
        proxy_send_timeout 10s;
    }

    # Rate limiting for write endpoints (prevent abuse)
    location /task {
        limit_req zone=write burst=5 nodelay;
        proxy_pass http://kanban_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Deny access to hidden files
    location ~ /\. {
        deny all;
    }
}

# Rate limit zone (defined in http block of /etc/nginx/nginx.conf)
# limit_req_zone $binary_remote_addr zone=write:1m rate=2r/s;
```

### TLS Setup

```sh
# Option A: With domain name
sudo certbot --nginx -d YOUR_DOMAIN --non-interactive --agree-tos -m YOUR_EMAIL

# Option B: IP-only (self-signed -- for initial testing before domain is confirmed)
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/kanban-selfsigned.key \
  -out /etc/ssl/certs/kanban-selfsigned.crt \
  -subj "/CN=$(curl -s ifconfig.me)"
# Update nginx config to reference self-signed cert paths
```

**Note**: Let's Encrypt requires a domain name. If no domain is confirmed, start with self-signed certs for HTTPS (browsers will show a warning) or run HTTP-only initially and add TLS when a domain is available.

### Enable the Site

```sh
sudo ln -s /etc/nginx/sites-available/kanban-web /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx
```

---

## 5. systemd Service Files

### Blue Instance: `/etc/systemd/system/kanban-web-blue.service`

```ini
[Unit]
Description=kanban-web (blue instance)
After=network.target
Wants=network.target

[Service]
Type=simple
User=kanban
Group=kanban
WorkingDirectory=/var/kanban

ExecStart=/var/kanban/bin/kanban-web \
  --addr :8080 \
  --repo ${KANBAN_WEB_REPO} \
  --clone-path /var/kanban/repo \
  --sync-interval 60s

EnvironmentFile=/var/kanban/env/kanban-web.env

Restart=on-failure
RestartSec=5
StartLimitBurst=3
StartLimitIntervalSec=60

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/kanban
PrivateTmp=true

# Logging to journal
StandardOutput=journal
StandardError=journal
SyslogIdentifier=kanban-web-blue

[Install]
WantedBy=multi-user.target
```

### Green Instance: `/etc/systemd/system/kanban-web-green.service`

Identical to blue except:

```ini
Description=kanban-web (green instance)
ExecStart=/var/kanban/bin/kanban-web \
  --addr :8081 \
  ...
SyslogIdentifier=kanban-web-green
```

### Environment File: `/var/kanban/env/kanban-web.env`

```sh
KANBAN_WEB_REPO=https://github.com/USER/REPO.git
KANBAN_WEB_COOKIE_KEY=<32-byte-hex-key>
# Generate key: openssl rand -hex 32
```

### Enable and Start

```sh
sudo systemctl daemon-reload
sudo systemctl enable kanban-web-blue
sudo systemctl start kanban-web-blue
# Green stays stopped until deployment activates it
```

---

## 6. Directory Layout on VM

```
/var/kanban/
  bin/
    kanban-web           # Active binary (symlink to blue or green version)
    kanban-web-blue      # Blue version binary
    kanban-web-green     # Green version binary
  repo/                  # Shared local git clone
  env/
    kanban-web.env       # Environment variables (600 permissions)
  .ssh/
    id_ed25519           # Deploy key (read-only)
    id_ed25519.pub
```

File permissions:

```sh
sudo chmod 600 /var/kanban/env/kanban-web.env
sudo chmod 700 /var/kanban/.ssh
sudo chmod 600 /var/kanban/.ssh/id_ed25519
```

---

## 7. Initial Git Clone

```sh
# Clone the repo as the kanban user
sudo -u kanban git clone https://github.com/USER/REPO.git /var/kanban/repo

# For private repos using deploy key:
sudo -u kanban git clone git@github.com:USER/REPO.git /var/kanban/repo
```

The background sync goroutine in `kanban-web` handles subsequent pulls automatically.

---

## 8. Rejected Simpler Alternatives

### Alternative 1: Cloud Run (Serverless)

- **What**: Deploy kanban-web as a Cloud Run service. No VM management.
- **Expected Impact**: Meets 90% of requirements (HTTP serving, scaling to zero).
- **Why Insufficient**: Cloud Run is stateless -- no persistent local git clone. Every request would need to clone/pull from GitHub, adding 2-5 seconds latency. The sync-interval approach requires a long-running process. Also, Cloud Run free tier has cold start latency that degrades user experience.

### Alternative 2: Single Instance (No Blue/Green)

- **What**: Run one kanban-web instance. Deploy by stopping, replacing binary, starting.
- **Expected Impact**: Meets 95% of requirements.
- **Why Insufficient**: User explicitly requested blue/green deployments (DD-07). Downtime during deploys is 5-15 seconds. For a student project this is acceptable, but the user wants to learn blue/green deployment patterns. The additional complexity is two systemd services and an nginx upstream toggle -- minimal overhead.
