# Deploy BookLib Backend to Railway.app

## 🚀 5-Minute Deployment Guide

Railway.app offers **$5 free credit per month** which covers this app completely ($0 out of pocket!).

### Prerequisites
- GitHub account with `booklib-backend` repo
- Railway.app account ([sign up free](https://railway.app))

---

## Step 1: Install Railway CLI

```bash
npm install -g @railway/cli
```

---

## Step 2: Login to Railway

```bash
railway login
```

This opens your browser to authenticate with GitHub.

---

## Step 3: Deploy from Your Repo

Navigate to your backend folder:
```bash
cd /Users/brandon/personal/booklib-backend
```

Initialize Railway project:
```bash
railway init
```

You'll be prompted:
- **Project name**: `booklib-backend` (or any name you prefer)
- **Environment**: `production`

Railway will create a new project and link it to your current directory.

---

## Step 4: Add Persistent Volume

**Option A: Via Dashboard (Recommended)**

```bash
# Open the Railway dashboard
railway open
```

Then in the dashboard:
1. Click on your service
2. Go to **Settings** tab
3. Scroll to **Volumes** section  
4. Click **+ New Volume**
5. Set mount path: `/data`
6. Set size: `1GB`
7. Click **Add**

**Option B: Via CLI**

Railway doesn't support volume creation via CLI yet, so use the dashboard method above.

---

## Step 5: Set Environment Variables

```bash
# Generate a secure session secret (used for both sessions and JWT tokens)
SESSION_SECRET=$(openssl rand -base64 32)

# Set all required variables
railway variables set \
  DATABASE_PATH=/data/booklib.db \
  PORT=8080 \
  SESSION_SECRET="$SESSION_SECRET" \
  CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174
```

**Note**: 
- `SESSION_SECRET` is used for JWT token signing and session management
- We'll update `CORS_ALLOWED_ORIGINS` after deploying the frontend
- You can optionally set `JWT_SECRET` separately if you want different keys for JWT vs sessions

---

## Step 6: Deploy!

```bash
railway up
```

Railway will:
1. Upload your code
2. Detect the Dockerfile
3. Build the Go application  
4. Deploy it
5. Assign a public URL

This takes about 2-3 minutes.

---

## Step 7: Get Your Backend URL

```bash
railway domain
```

Your backend will be at something like:
```
https://booklib-backend-production-xxxx.up.railway.app
```

**Save this URL!** You'll need it for the frontend.

---

## Step 8: Verify Deployment

Test the health check:
```bash
curl https://your-railway-url.up.railway.app/health
```

Expected response:
```json
{"status":"healthy","service":"booklib-backend"}
```

View logs:
```bash
railway logs
```

Look for:
```
Server starting on :8080
Database path: /data/booklib.db
Allowed CORS origins: [...]
```

---

## ✅ Backend Deployed!

Your backend is now live at your Railway URL! 🎉

**Cost**: $0/month (covered by $5 free credit - app uses ~$1.75/month)

---

## 🎨 Next Steps: Deploy Frontend

### 1. Update CORS

Once you deploy your frontend to Cloudflare Pages, update CORS:

```bash
railway variables set CORS_ALLOWED_ORIGINS=https://your-frontend-url.pages.dev
```

### 2. Deploy Frontend to Cloudflare Pages

See `/Users/brandon/personal/booklib-frontend/DEPLOYMENT.md` for frontend deployment instructions.

You'll need to:
1. Connect your GitHub repo to Cloudflare Pages
2. Set build command: `npm run build`
3. Set output directory: `dist`
4. Add environment variable: `VITE_API_URL=<your-railway-backend-url>`

---

## 🔧 Useful Commands

```bash
# View real-time logs
railway logs

# Check status
railway status

# Open dashboard
railway open

# Shell access (debugging)
railway shell

# List variables
railway variables

# Redeploy
railway up
```

---

## 💾 Backups

Railway provides volume snapshots:

1. Open dashboard: `railway open`
2. Go to your service → **Volumes**
3. Click on your volume
4. Click **"Create Snapshot"**

Snapshots are kept for 7 days on the free plan.

---

## 🆘 Troubleshooting

### Build Failed
```bash
# Check logs
railway logs

# Verify Dockerfile exists
ls -la Dockerfile
```

### Database Not Found
```bash
# Check volume is mounted
railway shell
df -h | grep /data
ls -lh /data
```

### CORS Errors
```bash
# Verify CORS setting
railway variables | grep CORS

# Update with frontend URL
railway variables set CORS_ALLOWED_ORIGINS=https://your-frontend.pages.dev
```

### Service Won't Start
```bash
# View detailed logs
railway logs --tail 100

# Check all environment variables
railway variables
```

---

## 📊 Monitor Usage

View your credit usage:
1. `railway open`
2. Click your profile → **Usage**
3. See how much of your $5 credit you're using

Expected: ~$1.75/month (well within free tier!)

---

## 🎉 Success!

Your backend is deployed and running on Railway for free! 

Next: Deploy your frontend to complete the full stack deployment.

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
