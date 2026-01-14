# /dev/dungeon Deployment Guide

Complete guide to deploying /dev/dungeon to Fly.io with the custom domain `dev-dungeon.com`.

---

## Prerequisites

- [ ] Fly.io account (free tier works): https://fly.io/
- [ ] Fly CLI installed
- [ ] Domain `dev-dungeon.com` with access to DNS settings
- [ ] ~10 minutes

### Install Fly CLI

```bash
# macOS
brew install flyctl

# Linux
curl -L https://fly.io/install.sh | sh

# Windows
powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
```

---

## Step 1: Login to Fly.io

```bash
fly auth login
```

This opens a browser to authenticate.

---

## Step 2: Create the App

```bash
fly apps create dev-dungeon
```

This creates an app named `dev-dungeon` on Fly.io.

---

## Step 3: Create Persistent Volume

The SSH host key must persist across deployments, otherwise users get "HOST IDENTIFICATION HAS CHANGED" warnings.

```bash
fly volumes create devdungeon_data --region ord --size 1
```

- `--region ord` = Chicago (change to your preferred region)
- `--size 1` = 1GB (minimum, only stores a tiny key file)

**Available regions:** Run `fly platform regions` to see all options.

---

## Step 4: Create PostgreSQL Database

```bash
# Create the database (this takes ~2 minutes)
fly postgres create --name dev-dungeon-db --region ord

# Attach it to your app (sets DATABASE_URL automatically)
fly postgres attach dev-dungeon-db --app dev-dungeon
```

Save the connection details shown - you'll need them if you want to connect directly.

---

## Step 5: Set Environment Variables (Secrets)

The `DATABASE_URL` is set automatically by `fly postgres attach`. You can verify with:

```bash
fly secrets list --app dev-dungeon
```

If you need to set it manually or add other secrets:

```bash
fly secrets set DATABASE_URL="postgres://user:password@host:5432/dbname?sslmode=require"
```

---

## Step 6: Deploy

```bash
fly deploy
```

This builds the Docker image and deploys it. First deploy takes ~3-5 minutes.

**Verify it's running:**

```bash
fly status
fly logs
```

You should see:
```
==> SSH host key not found at /data/host_key
==> Generating new ed25519 host key...
==> Host key generated successfully
==> Starting /dev/dungeon server...
```

---

## Step 7: Get Your IP Addresses

```bash
fly ips list
```

Output looks like:
```
VERSION  IP                      TYPE
v4       123.45.67.89            public
v6       2a09:8280:1::1:abc      public
```

Save these - you need them for DNS.

---

## Step 8: Add SSL Certificate

```bash
fly certs add dev-dungeon.com
fly certs add www.dev-dungeon.com
```

Check certificate status:
```bash
fly certs show dev-dungeon.com
```

---

## Step 9: Configure DNS

Go to your domain registrar (wherever you bought `dev-dungeon.com`) and add these records:

### Required Records

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | @ | `<your-fly-ipv4>` | 300 |
| AAAA | @ | `<your-fly-ipv6>` | 300 |
| CNAME | www | dev-dungeon.fly.dev | 300 |

### Example (if your IPs are 123.45.67.89 and 2a09:8280:1::1:abc)

| Type | Name | Value |
|------|------|-------|
| A | @ | 123.45.67.89 |
| AAAA | @ | 2a09:8280:1::1:abc |
| CNAME | www | dev-dungeon.fly.dev |

**DNS propagation takes 5-30 minutes** (sometimes up to 24 hours).

---

## Step 10: Verify Everything Works

### Test Web Portal

```bash
# Should return {"success":true,"data":{"status":"ok","service":"/dev/dungeon"}}
curl https://dev-dungeon.com/api/health
```

Or just visit https://dev-dungeon.com in your browser.

### Test SSH

```bash
# Connect (will prompt for registration if new key)
ssh player@dev-dungeon.com

# Verbose mode for debugging
ssh -v player@dev-dungeon.com
```

---

## Environment Variables Reference

### Required

| Variable | Description | Set By |
|----------|-------------|--------|
| `DATABASE_URL` | PostgreSQL connection string | `fly postgres attach` (automatic) |

### Optional (with defaults)

| Variable | Description | Default |
|----------|-------------|---------|
| `SSH_HOST` | SSH server bind address | `0.0.0.0` |
| `SSH_PORT` | SSH server port | `2222` |
| `HTTP_HOST` | HTTP server bind address | `0.0.0.0` |
| `HTTP_PORT` | HTTP server port | `8080` |
| `SSH_HOST_KEY_PATH` | Path to SSH host key | `/data/host_key` (on Fly.io) |

### Error Tracking (Sentry)

| Variable | Description | Default |
|----------|-------------|---------|
| `SENTRY_DSN` | Sentry DSN for error tracking | (disabled if not set) |
| `SENTRY_ENVIRONMENT` | Environment name (production, staging) | `development` |

To enable Sentry error tracking:

```bash
fly secrets set SENTRY_DSN="https://xxx@xxx.ingest.sentry.io/xxx" SENTRY_ENVIRONMENT="production"
```

### Setting Environment Variables

**For secrets (sensitive data like DATABASE_URL):**
```bash
fly secrets set VARIABLE_NAME="value"
```

**For non-sensitive config (in fly.toml):**
```toml
[env]
  SSH_PORT = "2222"
  HTTP_PORT = "8080"
```

---

## Ports & Services

| Service | Internal Port | External Port | Protocol |
|---------|---------------|---------------|----------|
| SSH | 2222 | 22 | TCP (raw) |
| HTTP | 8080 | 80 | HTTP |
| HTTPS | 8080 | 443 | HTTPS (TLS terminated by Fly) |

---

## Troubleshooting

### "Host key verification failed"

The SSH host key changed. This happens if the volume was deleted/recreated.

**For users:** Remove the old key:
```bash
ssh-keygen -R dev-dungeon.com
```

**To prevent:** Never delete the `devdungeon_data` volume.

### "Connection refused" on SSH

1. Check the app is running: `fly status`
2. Check logs: `fly logs`
3. Verify port 22 is exposed: Check `fly.toml` services section

### Database connection errors

1. Verify secret is set: `fly secrets list`
2. Check database is running: `fly postgres list`
3. Test connection: `fly postgres connect -a dev-dungeon-db`

### DNS not working

1. Check propagation: https://dnschecker.org/#A/dev-dungeon.com
2. Verify records are correct at registrar
3. Wait up to 24 hours (usually faster)

### Certificate errors

1. Check cert status: `fly certs show dev-dungeon.com`
2. DNS must be pointing to Fly.io first
3. Wait for automatic provisioning (~5 min after DNS propagates)

---

## Useful Commands

```bash
# View app status
fly status

# View logs (live)
fly logs

# View logs (recent)
fly logs --no-tail

# SSH into the container (for debugging)
fly ssh console

# Restart the app
fly apps restart dev-dungeon

# View secrets
fly secrets list

# Scale (if needed)
fly scale count 2  # Run 2 instances

# View current config
fly config show
```

---

## Database Management

### Connect to Postgres directly

```bash
fly postgres connect -a dev-dungeon-db
```

### Run migrations

Migrations run automatically on first boot via the schema in `migrations/001_initial_schema.sql`.

To run manually:
```bash
fly ssh console -a dev-dungeon
cat /app/migrations/001_initial_schema.sql | psql $DATABASE_URL
```

### Backup database

```bash
fly postgres backup create -a dev-dungeon-db
fly postgres backup list -a dev-dungeon-db
```

---

## Updating the App

After making code changes:

```bash
# Deploy new version
fly deploy

# Or deploy with a specific image tag
fly deploy --image-label v1.0.0
```

Zero-downtime deploys are automatic.

---

## Costs (Fly.io Free Tier)

| Resource | Free Allowance | This App Uses |
|----------|----------------|---------------|
| VMs | 3 shared-cpu-1x VMs | 1 VM |
| Memory | 256MB per VM | ~100MB |
| Storage | 3GB volumes | 1GB volume |
| Bandwidth | Unlimited (fair use) | Minimal |
| Postgres | 1 free instance | 1 instance |

**You should be within free tier** for light usage.

---

## Quick Reference Card

```bash
# Deploy
fly deploy

# Logs
fly logs

# Status
fly status

# SSH debug
ssh -v player@dev-dungeon.com

# Restart
fly apps restart dev-dungeon

# Database shell
fly postgres connect -a dev-dungeon-db
```

---

## After Deployment Checklist

- [ ] Web portal loads at https://dev-dungeon.com
- [ ] API health check returns OK: `curl https://dev-dungeon.com/api/health`
- [ ] SSH connection works: `ssh test@dev-dungeon.com`
- [ ] Registration flow works (connect with new SSH key)
- [ ] Game starts and saves progress
- [ ] Leaderboard shows at https://dev-dungeon.com/leaderboard

---

**Questions?** Check the logs first: `fly logs`
