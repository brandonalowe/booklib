# 🚀 BookLib Deployment Guide

## Quick Deploy to Railway.app (FREE - Recommended)

Railway provides $5 free credit per month (app uses ~$1.75/month = **$0 cost**).

### Step-by-Step Guide

**See [RAILWAY_QUICKSTART.md](./RAILWAY_QUICKSTART.md) for complete instructions.**

### TL;DR

```bash
npm install -g @railway/cli
railway login
cd booklib-backend
railway init
railway up
```

Then in Railway dashboard:
1. Add persistent volume: `/data` (1GB)
2. Set environment variables (see below)
3. Done! 🎉

---

## Environment Variables

Required variables for production:

```bash
DATABASE_PATH=/data/booklib.db
PORT=8080
SESSION_SECRET=<generate-with-openssl-rand>
CORS_ALLOWED_ORIGINS=https://your-frontend-url.pages.dev
```

Generate secure secret:
```bash
openssl rand -base64 32
```

---

## Alternative Platforms

| Platform | Cost | Always-On | Setup | Guide |
|----------|------|-----------|-------|-------|
| **Railway** | **$0** ⭐ | ✅ | Easy | [RAILWAY_QUICKSTART.md](./RAILWAY_QUICKSTART.md) |
| **Fly.io** | ~$3/mo | ✅ | Medium | Docs below |
| **Render** | $0* | ❌ Cold starts | Easy | Dashboard deploy |
| **Koyeb** | $0 | ✅ | Easy | Dashboard deploy |

*Render free tier spins down after 15 min inactivity (50s cold start)

---

## Fly.io Deployment (Paid)

**Cost**: ~$2.50-3/month

```bash
# Install CLI
brew install flyctl

# Login
flyctl auth login

# Create app
flyctl apps create booklib-backend

# Create volume
flyctl volumes create booklib_data --size 1 --region sjc

# Set secrets
flyctl secrets set SESSION_SECRET=$(openssl rand -base64 32)
flyctl secrets set CORS_ALLOWED_ORIGINS=https://your-frontend.pages.dev

# Deploy
flyctl deploy
```

Configuration files included: `fly.toml`, `Dockerfile`

---

## Frontend Deployment

Deploy frontend to **Cloudflare Pages** (FREE):

1. Connect GitHub repo to [Cloudflare Pages](https://pages.cloudflare.com/)
2. Build settings:
   - Command: `npm run build`
   - Output: `dist`
   - Node: `20`
3. Environment variable:
   - `VITE_API_URL` = `<your-railway-backend-url>`
4. Deploy!

Then update backend CORS:
```bash
railway variables set CORS_ALLOWED_ORIGINS=https://your-frontend.pages.dev
```

---

## Docker Deployment

```bash
docker build -t booklib-backend .

docker run -p 8080:8080 \
  -e SESSION_SECRET="your-secret" \
  -e DATABASE_PATH="/data/booklib.db" \
  -v booklib-data:/data \
  booklib-backend
```

---

## Health Check

Test your deployment:
```bash
curl https://your-backend-url.up.railway.app/health
```

Expected response:
```json
{"status":"healthy","service":"booklib-backend"}
```

---

## Troubleshooting

### Build Fails
- Check Go version in Dockerfile matches dependencies
- Verify all environment variables are set
- View logs: `railway logs`

### Database Not Persisting
- Ensure persistent volume is mounted at `/data`
- Check `DATABASE_PATH=/data/booklib.db`

### CORS Errors
- Verify `CORS_ALLOWED_ORIGINS` includes your frontend URL
- Update: `railway variables set CORS_ALLOWED_ORIGINS=<url>`

---

## Support

- **Full Railway Guide**: [RAILWAY_QUICKSTART.md](./RAILWAY_QUICKSTART.md)
- **Backend README**: [README.md](./README.md)
- **Frontend**: [booklib-frontend](https://github.com/brandonalowe/booklib-frontend)

---

Made with ❤️ for book lovers 📚
