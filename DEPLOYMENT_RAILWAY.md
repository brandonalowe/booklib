# Railway.app Deployment Guide (FREE Alternative)

Railway.app offers $5/month in free credits, which is enough to run BookLib for free!

## 🎯 Why Railway.app?

- ✅ **$5 free credit per month** (more than enough for this app)
- ✅ **Always-on** (doesn't sleep like Render.com)
- ✅ **Persistent volumes** included
- ✅ **Automatic HTTPS** and custom domains
- ✅ **Simple deployment** from GitHub
- ✅ **Great DX** - auto-detects Go apps
- ✅ **Built-in backups** with volume snapshots

## 🚀 Quick Deploy

### 1. Create Railway Account
Visit [railway.app](https://railway.app) and sign up with GitHub.

### 2. Create New Project

```bash
# Install Railway CLI (optional, but recommended)
npm install -g @railway/cli

# Login
railway login

# Link your project
cd booklib-backend
railway init
```

### 3. Deploy from Dashboard (Easier Method)

1. Go to [railway.app/new](https://railway.app/new)
2. Click **"Deploy from GitHub repo"**
3. Select `booklib-backend` repository
4. Railway will auto-detect it's a Go app!

### 4. Configure Environment Variables

In Railway dashboard, add these variables:

```bash
# Database path (Railway provides persistent disk at /data)
DATABASE_PATH=/data/booklib.db

# Server port (Railway provides PORT automatically, but we'll use 8080)
PORT=8080

# Session secret
SESSION_SECRET=<generate with: openssl rand -base64 32>

# CORS allowed origins (add your frontend URL)
CORS_ALLOWED_ORIGINS=https://booklib-frontend.pages.dev

# Optional: Email reminders
REMINDER_CRON_SCHEDULE=0 9 * * *
RUN_REMINDERS_ON_STARTUP=false
```

### 5. Add Persistent Volume

1. In your Railway service, go to **Settings**
2. Scroll to **Volumes**
3. Click **Add Volume**
4. Mount path: `/data`
5. Size: `1GB` (plenty for SQLite)

### 6. Deploy!

Railway will automatically:
- Build your Go app
- Deploy it
- Assign a public URL
- Enable HTTPS

Your backend will be at: `https://booklib-backend.up.railway.app`

## 🔄 CI/CD with Railway

Railway automatically deploys when you push to your main branch!

### Manual Deploy
```bash
railway up
```

### View Logs
```bash
railway logs
```

### Check Status
```bash
railway status
```

## 💾 Backups on Railway

Railway provides volume snapshots:

1. Go to your service → **Volumes**
2. Click on your volume
3. Use **"Create Snapshot"** button
4. Snapshots are stored for 7 days (on free plan)

### Automated Backups via Cron

Railway doesn't have built-in cron jobs, but you can:

**Option 1: Use GitHub Actions** (Recommended)
The `.github/workflows/backup.yml` file can be adapted:

```yaml
name: Railway Database Backup

on:
  schedule:
    - cron: '0 2 * * *'  # 2 AM daily
  workflow_dispatch:

jobs:
  backup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Railway CLI
        run: npm install -g @railway/cli
      
      - name: Create Volume Snapshot
        run: |
          railway login --browserless
          railway environment production
          # Railway CLI commands for backup
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

**Option 2: Self-hosted cron service**
- Use GitHub Actions (free)
- Use EasyCron (free tier)
- Use cron-job.org (free)

## 🔧 Dockerfile for Railway

Railway works with the existing Dockerfile! No changes needed.

But you can optimize it:

```dockerfile
# Railway-optimized Dockerfile (optional)
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app
COPY --from=builder /app/server .
COPY scripts/ ./scripts/

# Railway provides $PORT, but we'll expose 8080
EXPOSE 8080

CMD ["./server"]
```

## 📊 Monitor Usage & Costs

1. Go to **Account Settings** → **Usage**
2. View your credit usage (out of $5)
3. Typical usage for this app: **~$1-2/month** (well within free tier!)

### Expected Usage:
- **Compute**: ~$1.50/month (256MB RAM, always-on)
- **Storage**: ~$0.25/month (1GB volume)
- **Network**: Free (generous bandwidth)
- **Total**: ~$1.75/month = **FREE** (covered by $5 credit)

## 🔐 Custom Domain (Optional)

1. Go to service **Settings**
2. Under **Domains**, click **Generate Domain**
3. Or add custom domain:
   - Add your domain (e.g., `api.booklib.com`)
   - Add CNAME record: `booklib-backend.up.railway.app`
   - Railway handles SSL automatically!

## 🆘 Troubleshooting

### View Logs
```bash
railway logs
```

### Check Environment Variables
```bash
railway variables
```

### Shell Access
```bash
railway shell
```

### Database Issues
```bash
# Check if database exists
railway shell
ls -lh /data
```

### Redeploy
```bash
railway up --detach
```

## 🔄 Migration from Fly.io to Railway

If you already deployed to Fly.io and want to switch:

1. **Export your database:**
   ```bash
   flyctl ssh console -C "sqlite3 /data/booklib.db .dump" > backup.sql
   ```

2. **Deploy to Railway** (follow steps above)

3. **Import your data:**
   ```bash
   railway shell
   sqlite3 /data/booklib.db < backup.sql
   ```

## 📝 Update Frontend

Update your frontend environment variable:

**Cloudflare Pages:**
- Change `VITE_API_URL` to: `https://booklib-backend.up.railway.app`

**GitHub Secret:**
- Update `VITE_API_URL` in repository secrets

## ✅ Advantages over Fly.io

| Feature | Railway.app | Fly.io |
|---------|------------|--------|
| **Cost** | $0 (free $5 credit) | ~$2.50-3/month |
| **Sleep/Wake** | Always-on | Pay per hour |
| **Setup** | Simpler | More config |
| **Volumes** | Included | Pay extra |
| **Dashboard** | Modern UI | CLI-focused |
| **Cron Jobs** | Via GitHub Actions | Built-in |

## 🎉 You're Done!

Your app is now running on Railway for **FREE**! 🎊

- Backend: `https://booklib-backend.up.railway.app`
- Cost: **$0/month** (covered by free credit)
- Always-on, no cold starts
- Automatic deployments from GitHub
- Persistent SQLite database

Enjoy your free hosting! 🚀
