# 🚀 Deployment Guide - NAT Management System

## Table of Contents

- [Pre-Deployment Checklist](#pre-deployment-checklist)
- [Production Environment Setup](#production-environment-setup)
- [Deployment Methods](#deployment-methods)
- [Security Hardening](#security-hardening)
- [Performance Optimization](#performance-optimization)
- [Monitoring & Logging](#monitoring--logging)
- [Backup & Recovery](#backup--recovery)
- [Troubleshooting Production Issues](#troubleshooting-production-issues)

---

## Pre-Deployment Checklist

### ✅ Code Preparation

- [ ] All tests passing (`go test ./...`)
- [ ] Code reviewed and approved
- [ ] No debug/development code left
- [ ] Version number updated
- [ ] CHANGELOG.md updated
- [ ] Documentation up-to-date

### ✅ Configuration

- [ ] Production `.env` file prepared
- [ ] Strong JWT secret generated (min 32 chars)
- [ ] Strong session secret generated
- [ ] Database connection string configured
- [ ] CORS origins whitelist configured
- [ ] Rate limits set appropriately
- [ ] DEBUG mode disabled (`DEBUG=false`)

### ✅ Security

- [ ] Default passwords changed
- [ ] API keys rotated
- [ ] SSL/TLS certificates obtained
- [ ] Firewall rules configured
- [ ] Database backups enabled
- [ ] Security audit completed

### ✅ Infrastructure

- [ ] Production server provisioned
- [ ] PostgreSQL database ready (or Supabase project)
- [ ] Domain name configured
- [ ] Reverse proxy setup (Nginx/Caddy)
- [ ] Monitoring tools installed
- [ ] Backup system configured

---

## Production Environment Setup

### Server Requirements

**Minimum Specifications:**
- **CPU**: 2 cores
- **RAM**: 2 GB
- **Storage**: 20 GB SSD
- **OS**: Linux (Ubuntu 22.04 LTS recommended)
- **Network**: Stable connection to routers

**Recommended Specifications:**
- **CPU**: 4+ cores
- **RAM**: 4+ GB
- **Storage**: 50+ GB SSD
- **OS**: Linux (Ubuntu 22.04 LTS)
- **Network**: Redundant connections

### System Preparation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y git curl wget nginx postgresql-client

# Install Go 1.24
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Verify installation
go version
```

### Create Application User

```bash
# Create dedicated user (don't run as root!)
sudo useradd -r -m -s /bin/bash natapp
sudo mkdir -p /opt/nat-management
sudo chown natapp:natapp /opt/nat-management
```

### Clone and Build

```bash
# Switch to app user
sudo su - natapp

# Clone repository
cd /opt/nat-management
git clone <repository-url> .

# Install dependencies
go mod download

# Build production binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o nat-supabase \
  ./cmd/main.go

# Verify binary
./nat-supabase --version
```

### Production Environment File

Create `/opt/nat-management/.env`:

```env
# Production Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DEBUG=false

# Database (Supabase PostgreSQL)
DATABASE_URL=postgresql://postgres:password@db.xxxxx.supabase.co:5432/postgres?sslmode=require

# JWT Configuration (GENERATE STRONG SECRETS!)
JWT_SECRET=<GENERATE-STRONG-RANDOM-32+CHARS-SECRET>
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# Session Configuration (GENERATE STRONG SECRETS!)
SESSION_SECRET=<GENERATE-STRONG-RANDOM-32+CHARS-SECRET>
SESSION_MAX_AGE=86400

# CORS (Production domains only)
ALLOWED_ORIGINS=https://nat.example.com,https://www.nat.example.com

# Rate Limiting (Production values)
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=60s

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

**Generate Strong Secrets:**
```bash
# Generate JWT secret
openssl rand -base64 48

# Generate session secret
openssl rand -base64 48
```

**Set Permissions:**
```bash
chmod 600 /opt/nat-management/.env
chown natapp:natapp /opt/nat-management/.env
```

---

## Deployment Methods

### Method 1: Systemd Service (Recommended)

Create `/etc/systemd/system/nat-management.service`:

```ini
[Unit]
Description=NAT Management Application
After=network.target

[Service]
Type=simple
User=natapp
Group=natapp
WorkingDirectory=/opt/nat-management
ExecStart=/opt/nat-management/nat-supabase
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=nat-management

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/nat-management

# Resource limits
LimitNOFILE=65536
LimitNPROC=512

[Install]
WantedBy=multi-user.target
```

**Enable and Start Service:**
```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service (start on boot)
sudo systemctl enable nat-management

# Start service
sudo systemctl start nat-management

# Check status
sudo systemctl status nat-management

# View logs
sudo journalctl -u nat-management -f
```

**Service Management:**
```bash
# Stop service
sudo systemctl stop nat-management

# Restart service
sudo systemctl restart nat-management

# Reload configuration (without restart)
sudo systemctl reload nat-management
```

### Method 2: Docker Deployment

**Dockerfile:**

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o nat-supabase ./cmd/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/nat-supabase .
COPY --from=builder /app/web ./web
COPY .env .

EXPOSE 8080

USER 1000:1000

CMD ["./nat-supabase"]
```

**docker-compose.yml:**

```yaml
version: '3.8'

services:
  nat-management:
    build: .
    container_name: nat-management
    restart: always
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - JWT_SECRET=${JWT_SECRET}
      - SESSION_SECRET=${SESSION_SECRET}
    env_file:
      - .env
    volumes:
      - ./web:/app/web:ro
    networks:
      - nat-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

networks:
  nat-network:
    driver: bridge
```

**Deploy with Docker:**
```bash
# Build image
docker-compose build

# Start container
docker-compose up -d

# View logs
docker-compose logs -f

# Stop container
docker-compose down
```

### Method 3: Binary Deployment

Simple deployment without systemd or Docker:

```bash
# Run in background with nohup
nohup /opt/nat-management/nat-supabase > /var/log/nat-management.log 2>&1 &

# Or use screen
screen -S nat-management
/opt/nat-management/nat-supabase
# Press Ctrl+A, then D to detach

# Reattach later
screen -r nat-management
```

---

## Security Hardening

### 1. Reverse Proxy Setup (Nginx)

**Install Nginx:**
```bash
sudo apt install nginx
```

**Configure Nginx** (`/etc/nginx/sites-available/nat-management`):

```nginx
# Rate limiting
limit_req_zone $binary_remote_addr zone=nat_limit:10m rate=10r/s;

# Upstream
upstream nat_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name nat.example.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name nat.example.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/nat.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/nat.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Logging
    access_log /var/log/nginx/nat-management-access.log;
    error_log /var/log/nginx/nat-management-error.log;

    # Rate limiting
    limit_req zone=nat_limit burst=20 nodelay;

    # Proxy settings
    location / {
        proxy_pass http://nat_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Static files
    location /static/ {
        alias /opt/nat-management/web/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

**Enable Site:**
```bash
sudo ln -s /etc/nginx/sites-available/nat-management /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 2. SSL/TLS Certificate (Let's Encrypt)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d nat.example.com

# Auto-renewal
sudo certbot renew --dry-run
```

### 3. Firewall Configuration (UFW)

```bash
# Enable UFW
sudo ufw enable

# Allow SSH (change port if non-standard)
sudo ufw allow 22/tcp

# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Deny direct access to app port
sudo ufw deny 8080/tcp

# Check status
sudo ufw status
```

### 4. Database Security

**For Supabase:**
- Use connection pooling (PgBouncer built-in)
- Enable SSL (sslmode=require)
- Use strong passwords
- Regularly rotate credentials
- Configure Row Level Security (RLS) policies
- Use Supabase dashboard for backups

**For Self-hosted PostgreSQL:**
```bash
# Configure pg_hba.conf
# Only allow connections from application server
host    nat_management    natapp    <app-server-ip>/32    md5

# Restart PostgreSQL
sudo systemctl restart postgresql
```

### 5. Application Security

**Environment Variables:**
```bash
# Restrict access to .env
chmod 600 /opt/nat-management/.env
chown natapp:natapp /opt/nat-management/.env
```

**File Permissions:**
```bash
# Set correct ownership
sudo chown -R natapp:natapp /opt/nat-management

# Binary permissions
chmod 755 /opt/nat-management/nat-supabase

# Web files (read-only)
chmod -R 644 /opt/nat-management/web
find /opt/nat-management/web -type d -exec chmod 755 {} \;
```

---

## Performance Optimization

### 1. Database Optimization

**Connection Pooling** (automatic with pgx):
```env
# In .env
DATABASE_URL=postgresql://user:pass@host/db?pool_max_conns=10&pool_min_conns=2
```

**Database Indexes:**
Already included in migrations, but verify:
```sql
-- Critical indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_routers_name ON routers(name);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created ON activity_logs(created_at DESC);
```

### 2. Application Optimization

**Build Optimizations:**
```bash
# Strip debug symbols and reduce binary size
go build -ldflags="-w -s" -o nat-supabase ./cmd/main.go

# Enable all optimizations
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -trimpath \
  -o nat-supabase \
  ./cmd/main.go
```

**Runtime Settings:**
```bash
# Set GOMAXPROCS (defaults to CPU count)
export GOMAXPROCS=4

# Increase file descriptor limit
ulimit -n 65536
```

### 3. Nginx Caching

Add to Nginx config:
```nginx
# Cache zone
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=nat_cache:10m max_size=100m inactive=60m;

location /api/ {
    # Cache GET requests for 30 seconds
    proxy_cache nat_cache;
    proxy_cache_valid 200 30s;
    proxy_cache_key "$scheme$request_method$host$request_uri";

    proxy_pass http://nat_backend;
}
```

### 4. System Tuning

**Kernel Parameters** (`/etc/sysctl.conf`):
```conf
# Network performance
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535

# Connection tracking
net.netfilter.nf_conntrack_max = 524288

# Apply changes
# sudo sysctl -p
```

---

## Monitoring & Logging

### 1. Application Logging

**Structured Logging:**
Application already uses logrus with structured logging.

**Log Rotation:**
```bash
# Create logrotate config
sudo nano /etc/logrotate.d/nat-management
```

Content:
```
/var/log/nat-management/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 natapp natapp
    sharedscripts
    postrotate
        systemctl reload nat-management > /dev/null 2>&1 || true
    endscript
}
```

### 2. System Monitoring

**Monitor with systemctl:**
```bash
# Service status
sudo systemctl status nat-management

# Recent logs
sudo journalctl -u nat-management -n 100

# Follow logs
sudo journalctl -u nat-management -f

# Logs with timestamp
sudo journalctl -u nat-management --since "1 hour ago"
```

### 3. Resource Monitoring

**Install monitoring tools:**
```bash
sudo apt install htop iotop nethogs
```

**Monitor resources:**
```bash
# CPU/Memory usage
htop

# Disk I/O
sudo iotop

# Network bandwidth
sudo nethogs

# Application-specific
ps aux | grep nat-supabase
```

### 4. Health Checks

**Create health check endpoint:**
Already available at `/api/auth/check`

**External Monitoring:**
Use services like:
- UptimeRobot (free)
- Pingdom
- StatusCake
- Datadog

### 5. Alerting

**Email Alerts (Systemd):**
```bash
# Install mail utilities
sudo apt install mailutils

# Configure systemd to send email on failure
sudo systemctl edit nat-management
```

Add:
```ini
[Service]
OnFailure=failure-notification@%n.service
```

---

## Backup & Recovery

### 1. Database Backup

**Automated PostgreSQL Backup:**
```bash
#!/bin/bash
# /opt/nat-management/scripts/backup-db.sh

BACKUP_DIR="/opt/backups/nat-management"
DATE=$(date +%Y%m%d_%H%M%S)
DB_URL="postgresql://user:pass@host/nat_management"

mkdir -p $BACKUP_DIR

# Backup
pg_dump $DB_URL | gzip > "$BACKUP_DIR/backup_$DATE.sql.gz"

# Keep only last 30 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete

echo "Backup completed: backup_$DATE.sql.gz"
```

**Schedule with Cron:**
```bash
crontab -e
```

Add:
```
# Daily backup at 2 AM
0 2 * * * /opt/nat-management/scripts/backup-db.sh >> /var/log/nat-backup.log 2>&1
```

**For Supabase (Serverless):**
Supabase provides automatic backups. Check Supabase dashboard for backup options.

### 2. Application Backup

```bash
#!/bin/bash
# /opt/nat-management/scripts/backup-app.sh

BACKUP_DIR="/opt/backups/nat-management"
DATE=$(date +%Y%m%d_%H%M%S)
APP_DIR="/opt/nat-management"

mkdir -p $BACKUP_DIR

# Backup application files (exclude logs, temp files)
tar -czf "$BACKUP_DIR/app_$DATE.tar.gz" \
    --exclude='*.log' \
    --exclude='tmp/*' \
    -C /opt nat-management

echo "Application backup completed: app_$DATE.tar.gz"
```

### 3. Configuration Backup

**Backup .env and configs:**
```bash
# Encrypted backup of secrets
tar -czf - /opt/nat-management/.env /etc/nginx/sites-available/nat-management | \
    openssl enc -aes-256-cbc -salt -pbkdf2 -out backup_config_$(date +%Y%m%d).tar.gz.enc
```

### 4. Recovery Procedure

**Restore Database:**
```bash
# Extract backup
gunzip backup_20251016_020000.sql.gz

# Restore
psql $DATABASE_URL < backup_20251016_020000.sql
```

**Restore Application:**
```bash
# Stop service
sudo systemctl stop nat-management

# Extract backup
sudo tar -xzf app_20251016_020000.tar.gz -C /opt/

# Start service
sudo systemctl start nat-management
```

---

## Troubleshooting Production Issues

### Common Issues

#### 1. Application Won't Start

**Check logs:**
```bash
sudo journalctl -u nat-management -n 50
```

**Common causes:**
- Database connection failed → Check DATABASE_URL
- Port already in use → Check `sudo lsof -i:8080`
- Permission denied → Check file ownership/permissions

#### 2. High Memory Usage

```bash
# Check memory usage
ps aux | grep nat-supabase

# If too high, restart service
sudo systemctl restart nat-management
```

#### 3. Database Connection Issues

```bash
# Test database connection
psql $DATABASE_URL -c "SELECT 1;"

# Check connection pool
# Monitor active connections in database
```

#### 4. Router Connection Timeouts

```bash
# Use diagnostic tool
cd /opt/nat-management/tools
./router-diagnostic.exe <router-ip> 8728 <user> <pass>

# Check firewall
sudo ufw status
```

### Emergency Procedures

**Rollback Deployment:**
```bash
# Stop current version
sudo systemctl stop nat-management

# Restore previous backup
sudo tar -xzf /opt/backups/nat-management/app_PREVIOUS.tar.gz -C /opt/

# Start service
sudo systemctl start nat-management
```

**Database Recovery:**
```bash
# Restore from backup
psql $DATABASE_URL < /opt/backups/nat-management/backup_LATEST.sql
```

---

## Production Checklist

### Pre-Launch

- [ ] All tests passed
- [ ] Security audit completed
- [ ] SSL certificates installed
- [ ] Backups configured and tested
- [ ] Monitoring set up
- [ ] Documentation reviewed
- [ ] Team trained on procedures

### Post-Launch

- [ ] Monitor logs for errors
- [ ] Check performance metrics
- [ ] Verify backups working
- [ ] Test disaster recovery
- [ ] Monitor resource usage
- [ ] Collect user feedback

### Ongoing Maintenance

- [ ] Regular security updates
- [ ] Database optimization
- [ ] Log rotation working
- [ ] Backups verified monthly
- [ ] Performance monitoring
- [ ] Capacity planning

---

**Version:** 4.1
**Last Updated:** 2025-10-16
**Support:** See [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
