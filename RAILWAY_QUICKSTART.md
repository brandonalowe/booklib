# Railway.app Quick Start Guide

## 🚀 Deploy BookLib Backend to Railway (5 Steps)

### Step 1: Sign Up & Install CLI

```bash
# Sign up at railway.app (use GitHub for easy access)
open https://railway.app

# Install Railway CLI
npm install -g @railway/cli

# Login
railway login
```

### Step 2: Create Project

**Option A: Via CLI (Recommended)**
```bash
cd /Users/brandon/personal/booklib-backend

# Initialize Railway project
railway init

# Follow prompts:
# - Project name: booklib-backend
# - Environment: production
```

**Option B: Via Dashboard**
1. Go to [railway.app/new](https://railway.app/new)
2. Click **"Deploy from GitHub repo"**
3. Select `booklib-backend` repository
4. Railway auto-detects it's a Go app!

### Step 3: Add Persistent Volume

**Via Dashboard (easier):**
1. Go to your service settings
2. Click **"Volumes"** tab
3. Click **"+ New Volume"**
4. Mount path: `/data`
5. Size: `1GB` (plenty for SQLite)

**Via CLI:**
```bash
railway volume create --mount-path /data --size 1
```

### Step 4: Set Environment Variables

```bash
# Generate a secure session secret
export SESSION_SECRET=$(openssl rand -base64 32)

# Add environment variables
railway variables set DATABASE_PATH=/data/booklib.db
railway variables set PORT=8080
railway variables set SESSION_SECRET="$SESSION_SECRET"
railway variables set CORS_ALLOWED_ORIGINS=https://booklib-frontend.pages.dev

# Optional: Email reminders
railway variables set REMINDER_CRON_SCHEDULE="0 9 * * *"
railway variables set RUN_REMINDERS_ON_STARTUP=false
```

**Or via Dashboard:**
1. Go to your service → **Variables** tab
2. Add each variable:
   - `DATABASE_PATH` = `/data/booklib.db`
   - `PORT` = `8080`
   - `SESSION_SECRET` = (generate with `openssl rand -base64 32`)
   - `CORS_ALLOWED_ORIGINS` = `https://booklib-frontend.pages.dev`

### Step 5: Deploy!

```bash
# Deploy from CLI
railway up

# Or just push to GitHub (auto-deploys)
git push origin main
```

That's it! 🎉

---

## 📝 Your Backend URL

After deployment, Railway will give you a URL like:
```
https://booklib-backend.up.railway.app
```

**Find it:**
- Dashboard: Service → **Settings** → **Domains**
- CLI: `railway domain`

**Save this URL** - you'll need it for the frontend!

---

## 🔧 Useful Commands

```bash
# View logs in real-time
railway logs

# Check deployment status
railway status

# Open dashboard
railway open

# Shell access (for debugging)
railway shell

# List environment variables
railway variables

# Redeploy
railway up

# View service info
railway info
```

---

## ✅ Verify Deployment

### 1. Check Health
```bash
# Replace with your actual URL
curl https://booklib-backend.up.railway.app/health
```

Should return: `{"status":"healthy","service":"booklib-backend"}`

### 2. Check Logs
```bash
railway logs
```

Look for:
```
Server starting on :8080
Database path: /data/booklib.db
Allowed CORS origins: [...]
```

### 3. Test Database
```bash
railway shell
ls -lh /data
# Should see booklib.db after first request
```

---

## 🎨 Next: Deploy Frontend to Cloudflare Pages

### Update Frontend Configuration

1. **Get your Railway backend URL**: `railway domain`

2. **Add GitHub Secret**:
   - Go to GitHub repo settings → Secrets and variables → Actions
   - Add secret: `VITE_API_URL` = `https://booklib-backend.up.railway.app`

3. **Deploy to Cloudflare Pages**:
   - Go to [dash.cloudflare.com/pages](https://dash.cloudflare.com/pages)
   - Click **"Create a project"**
   - Connect GitHub repo: `booklib-frontend`
   - Build settings:
     - Build command: `npm run build`
     - Build output directory: `dist`
     - Node version: `20`
   - Environment variable:
     - `VITE_API_URL` = `https://booklib-backend.up.railway.app`
   - Click **"Save and Deploy"**

4. **Update CORS** (after frontend deploys):
```bash
# Get your Cloudflare Pages URL (e.g., https://booklib-frontend.pages.dev)
railway variables set CORS_ALLOWED_ORIGINS=https://booklib-frontend.pages.dev
```

---

## 💾 Backups on Railway

Railway provides **volume snapshots**:

### Create Manual Snapshot
1. Dashboard → Service → **Volumes**
2. Click on your volume
3. Click **"Create Snapshot"**

### GitHub Actions for Automated Backups

Add Railway token to GitHub secrets:
```bash
# Get token
railway token

# Add to GitHub secrets as RAILWAY_TOKEN
```

Then the existing `.github/workflows/backup.yml` will work (you'll need to adapt it for Railway).

**Simpler option**: Use Railway's built-in snapshots manually (free plan keeps for 7 days).

---

## 🔐 Custom Domain (Optional)

1. Go to service **Settings** → **Domains**
2. Click **"Custom Domain"**
3. Add: `api.yourdomain.com`
4. Add CNAME record in your DNS:
   ```
   api.yourdomain.com → booklib-backend.up.railway.app
   ```
5. Railway handles SSL automatically!

---

## 💰 Monitor Usage

1. Dashboard → **Account Settings** → **Usage**
2. Monitor your $5 credit usage
3. Expected usage: ~$1.75/month (well within free credit!)

---

## 🆘 Troubleshooting

### "Build failed"
```bash
# Check logs
railway logs

# Common fix: ensure Dockerfile exists
ls -la Dockerfile
```

### "Database not found"
```bash
# Check volume is mounted
railway shell
df -h
ls -lh /data
```

### "CORS errors in frontend"
```bash
# Verify CORS settings
railway variables

# Update if needed
railway variables set CORS_ALLOWED_ORIGINS=https://your-frontend-url.pages.dev
```

### "Service won't start"
```bash
# Check logs for errors
railway logs --tail 100

# Verify environment variables
railway variables
```

---

## 🎉 You're Done!

Your backend is now running on Railway for **FREE**!

- ✅ Always-on (no cold starts)
- ✅ Automatic deployments from GitHub
- ✅ Persistent SQLite database
- ✅ Free SSL & custom domains
- ✅ Covered by $5/month free credit

**Next**: Deploy your frontend to Cloudflare Pages following the steps above!

Need help? Check logs with `railway logs` or open the dashboard with `railway open`.
