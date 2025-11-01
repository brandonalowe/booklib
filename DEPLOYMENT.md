# BookLib Backend Deployment Guide

## 🎯 Recommended: Railway.app (FREE)

Railway.app is the recommended hosting platform for BookLib backend because:
- ✅ **$5 free credit per month** (app uses ~$1.75/month = $0 out of pocket)
- ✅ **Always-on** (no cold starts)
- ✅ **Simple setup** (auto-detects Go apps)
- ✅ **Persistent volumes** included
- ✅ **Auto-deploy** from GitHub

### Quick Start

See **[RAILWAY_QUICKSTART.md](./RAILWAY_QUICKSTART.md)** for step-by-step instructions.

**TL;DR:**
```bash
npm install -g @railway/cli
railway login
railway init
railway up
```

Then add a 1GB volume at `/data` via the dashboard.

---

## 📋 Prerequisites

## 📋 Prerequisites

**For Railway.app (Recommended):**
- GitHub account with `booklib-backend` repo
- Railway.app account ([sign up free](https://railway.app))
- Node.js installed (for Railway CLI)

**For Cloudflare Pages (Frontend):**
- Cloudflare account ([sign up free](https://dash.cloudflare.com/sign-up))

---

## 🚀 Backend Options

### Option 1: Railway.app (⭐ Recommended - FREE)

**Cost**: $0/month (covered by $5 free credit)  
**Guide**: [RAILWAY_QUICKSTART.md](./RAILWAY_QUICKSTART.md)

**Pros:**
- Completely free (with $5 monthly credit)
- Always-on, no cold starts
- Simple dashboard and CLI
- Automatic GitHub deployments
- Persistent volumes included

### Option 2: Fly.io (Paid)

### 1. Initialize Fly.io App

```bash
cd booklib-backend

# Login to Fly.io
flyctl auth login

# Create the app (choose your preferred region)
flyctl apps create booklib-backend

# Create persistent volume for database
flyctl volumes create booklib_data --size 1 --region sjc

# Create backup volume
flyctl volumes create booklib_backup --size 3 --region sjc
```

### 2. Set Environment Variables

```bash
# Set session secret (generate a random string)
flyctl secrets set SESSION_SECRET=$(openssl rand -base64 32)

# Set CORS allowed origins (use your Cloudflare Pages URL)
flyctl secrets set CORS_ALLOWED_ORIGINS=https://booklib-frontend.pages.dev
```

### 3. Deploy Backend

```bash
# Initial deployment
flyctl deploy

# Check status
flyctl status

# View logs
flyctl logs
```

### 4. Get Backend URL

After deployment, your backend will be available at:
```
https://booklib-backend.fly.dev
```

Save this URL - you'll need it for frontend configuration.

### 5. Setup Automated Backups

The backup system is already configured! To manually trigger a backup:

```bash
flyctl ssh console -C "/root/scripts/backup.sh"
```

To view backups:
```bash
flyctl ssh console -C "ls -lh /backup"
```

Backups run automatically at 2 AM UTC daily via GitHub Actions.

## 🎨 Frontend Deployment (Cloudflare Pages)

### 1. Get Cloudflare API Credentials

1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Navigate to **My Profile** → **API Tokens**
3. Click **Create Token** → Use **Edit Cloudflare Workers** template
4. Save your API token
5. Get your Account ID from the Workers dashboard

### 2. Configure GitHub Secrets

Add these secrets to your GitHub repository (Settings → Secrets and variables → Actions):

**Backend secrets:**
- `FLY_API_TOKEN`: Get from `flyctl auth token`

**Frontend secrets:**
- `CLOUDFLARE_API_TOKEN`: From step 1
- `CLOUDFLARE_ACCOUNT_ID`: From Cloudflare dashboard
- `VITE_API_URL`: Your backend URL (e.g., `https://booklib-backend.fly.dev`)

### 3. Deploy Frontend

```bash
cd booklib-frontend

# Create .env.local for local development
cp .env.local.example .env.local
# Edit .env.local to use http://localhost:8080
```

Push to GitHub, and the GitHub Action will automatically deploy to Cloudflare Pages!

Your frontend will be available at:
```
https://booklib-frontend.pages.dev
```

### 4. Custom Domain (Optional)

**For Cloudflare Pages:**
1. Go to your Cloudflare Pages project
2. Click **Custom domains** → **Set up a custom domain**
3. Add your domain (e.g., `booklib.yourdomain.com`)

**For Fly.io Backend:**
```bash
flyctl certs create api.yourdomain.com
```

Then add a CNAME record in your DNS:
```
api.yourdomain.com → booklib-backend.fly.dev
```

## 🔒 Update CORS Configuration

After getting your final frontend URL, update the backend CORS:

```bash
cd booklib-backend
flyctl secrets set CORS_ALLOWED_ORIGINS=https://your-custom-domain.com,https://booklib-frontend.pages.dev
```

## 📊 Monitoring & Maintenance

### View Backend Logs
```bash
flyctl logs -a booklib-backend
```

### View Metrics
```bash
flyctl metrics -a booklib-backend
```

### SSH into Backend
```bash
flyctl ssh console -a booklib-backend
```

### Manual Backup
```bash
flyctl ssh console -C "/root/scripts/backup.sh" -a booklib-backend
```

### Restore from Backup
```bash
# SSH into the machine
flyctl ssh console -a booklib-backend

# List backups
ls -lh /backup

# Stop the app temporarily
supervisorctl stop all

# Restore backup
cp /backup/booklib_20240101_120000.db /data/booklib.db

# Restart app
supervisorctl start all
```

## 💰 Cost Breakdown

⚠️ **Updated November 2025**: Fly.io no longer offers a free tier. Here's the expected cost:

- **Fly.io Pay-as-you-go:**
  - Shared CPU 1x (256MB RAM): ~$1.94/month (stopped when idle)
  - 1GB persistent storage: ~$0.15/month
  - 3GB backup storage: ~$0.45/month
  - Estimated total: **~$2.50-3/month** (with auto-stop enabled)

- **Cloudflare Pages Free Tier:**
  - Unlimited static requests
  - Unlimited bandwidth
  - 500 builds per month
  - 100 custom domains

**Total: ~$2.50-3/month** 💰

**Note**: See "Alternative Hosting Options" below for completely free alternatives.

## 🔄 CI/CD Workflow

Your deployments are now automated:

1. **Backend:** Push to `main` branch triggers deployment to Fly.io
2. **Frontend:** Push to `main` branch triggers build and deployment to Cloudflare Pages
3. **Backups:** Run daily at 2 AM UTC automatically

## 🛠️ Troubleshooting

### Backend won't start
```bash
# Check logs
flyctl logs -a booklib-backend

# Check if volume is mounted
flyctl ssh console -C "df -h" -a booklib-backend
```

### Database errors
```bash
# Verify database exists
flyctl ssh console -C "ls -lh /data" -a booklib-backend

# Check database integrity
flyctl ssh console -C "sqlite3 /data/booklib.db 'PRAGMA integrity_check;'" -a booklib-backend
```

### CORS errors in frontend
- Verify `CORS_ALLOWED_ORIGINS` includes your frontend URL
- Check browser console for exact error
- Update secrets: `flyctl secrets set CORS_ALLOWED_ORIGINS=...`

### Backup failures
```bash
# Manually run backup to see errors
flyctl ssh console -C "/root/scripts/backup.sh" -a booklib-backend

# Check backup volume
flyctl ssh console -C "df -h /backup" -a booklib-backend
```

## 📝 Environment Variables Reference

### Backend (Fly.io)
- `DATABASE_PATH`: `/data/booklib.db` (set in fly.toml)
- `PORT`: `8080` (set in fly.toml)
- `SESSION_SECRET`: Random string for session encryption (secret)
- `CORS_ALLOWED_ORIGINS`: Comma-separated frontend URLs (secret)

### Frontend (Cloudflare Pages)
- `VITE_API_URL`: Backend URL (e.g., `https://booklib-backend.fly.dev`)

## 🎉 Next Steps

1. Create a custom domain for a professional look
2. Set up error monitoring (e.g., Sentry)
3. Add Google Books API key for enhanced metadata
4. Configure email service for lending reminders
5. Implement the stats charts using Chart.js

---

**Need help?** Check the [Fly.io docs](https://fly.io/docs) or [Cloudflare Pages docs](https://developers.cloudflare.com/pages/).

---

## 🆓 Alternative Hosting Options (Completely Free)

If you want a $0/month solution, consider these alternatives:

### Option 1: Railway.app (Recommended Free Alternative)
- **Free tier**: $5 credit/month (typically covers a small app)
- **Includes**: PostgreSQL/MySQL database support
- **Pros**: Very easy deployment, generous free tier, automatic HTTPS
- **Setup**: Connect GitHub, Railway auto-detects Go app
- **Database**: Can use SQLite or upgrade to PostgreSQL (recommended)

### Option 2: Render.com
- **Free tier**: Available with limitations
- **Includes**: 750 hours/month free (enough for 1 service)
- **Pros**: Simple deployment, automatic SSL, disk persistence
- **Cons**: Services spin down after 15 min inactivity (50s cold start)
- **Good fit**: SQLite works, backups via cron jobs

### Option 3: Koyeb
- **Free tier**: 1 web service free forever
- **Includes**: 512MB RAM, 2GB storage
- **Pros**: Good for Go apps, stays running, persistent volumes
- **Cons**: Newer platform, smaller community

### Option 4: Self-host with Oracle Cloud (Advanced)
- **Free tier**: 2 VMs with 1GB RAM each, 200GB storage forever
- **Pros**: Most generous free tier, full control
- **Cons**: Requires server management, manual setup

### Recommended Free Stack:
**Railway.app + Cloudflare Pages**
- Backend: Railway.app (covered by $5/month credit)
- Frontend: Cloudflare Pages (unlimited free)
- Database: SQLite on Railway's persistent disk
- **Total: $0/month** ✨

Would you like me to create Railway.app deployment configs instead?
