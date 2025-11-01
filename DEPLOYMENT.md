# 🚀 Deploy to Railway

## Prerequisites
- GitHub account
- Railway account (free $5/month credit)

## Quick Deploy (10 minutes)

### 1. Connect GitHub to Railway

```bash
# Install Railway CLI (optional)
npm install -g @railway/cli
railway login
```

**Or use the dashboard:**
1. Go to https://railway.app/new
2. Click **Deploy from GitHub repo**
3. Select `booklib-backend`
4. Railway auto-detects Go and deploys!

### 2. Configure Environment Variables

In Railway dashboard → Your Project → Variables, add:

```bash
# Required
SESSION_SECRET=your-random-secret-here  # Generate: openssl rand -base64 32
PORT=8080

# Optional - CORS
CORS_ALLOWED_ORIGINS=https://your-frontend.pages.dev

# Optional - Email (see docs/EMAIL_SETUP_CUSTOM_DOMAIN.md)
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_USERNAME=resend
SMTP_PASSWORD=your-api-key
SMTP_FROM_EMAIL=no-reply@yourdomain.com
SMTP_FROM_NAME=BookLib
```

### 3. Add Persistent Volume

1. Railway dashboard → Your Project → **Settings**
2. Scroll to **Volumes**
3. Click **Add Volume**
4. Mount path: `/data`
5. Set `DATABASE_PATH=/data/booklib.db` in Variables

### 4. Deploy

Railway auto-deploys on push to `prod` branch:

```bash
git push origin prod
```

Your API is live at `https://your-app.up.railway.app`!

## Update CORS for Frontend

After frontend is deployed, update backend CORS:

```bash
CORS_ALLOWED_ORIGINS=https://your-frontend.pages.dev
```

## Database Backups

Railway provides automatic volume snapshots. Manual backup:

```bash
railway run bash scripts/backup.sh
```

## Troubleshooting

**"Database locked"**: Restart the service in Railway dashboard  
**CORS errors**: Verify `CORS_ALLOWED_ORIGINS` matches your frontend URL exactly  
**502 errors**: Check Railway logs for startup errors

## Custom Domain (Optional)

1. Railway → Settings → Domains
2. Add your domain
3. Update DNS with Railway's CNAME

---

**You're live!** 🎉 Test: `curl https://your-app.up.railway.app/health`
