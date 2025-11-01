# 📧 Email Setup with Custom Domain (booklib.app)

## Overview

This guide will help you set up email sending from your custom domain (`no-reply@booklib.app`) for sending book reminder emails.

**Recommended Solution**: [Resend](https://resend.com) - Modern email API built for developers.

---

## 🚀 Option 1: Resend (Recommended)

### Why Resend?
- ✅ **Free tier**: 3,000 emails/month, 100 emails/day
- ✅ **Simple setup**: Just add DNS records
- ✅ **Great deliverability**: Built by former Postmark team
- ✅ **SMTP support**: Works with your existing code
- ✅ **Custom domain**: Easy verification

### Step-by-Step Setup

#### 1️⃣ Create Resend Account

1. Go to https://resend.com
2. Sign up with GitHub or email
3. Verify your email

#### 2️⃣ Add Your Domain

1. In Resend dashboard, click **Domains**
2. Click **Add Domain**
3. Enter: `booklib.app`
4. Click **Add**

#### 3️⃣ Add DNS Records in Cloudflare

Resend will show you DNS records to add. Go to Cloudflare:

1. Login to https://dash.cloudflare.com
2. Select your domain: `booklib.app`
3. Go to **DNS** → **Records**
4. Add the following records (Resend provides these):

| Type | Name | Content | Proxy Status |
|------|------|---------|--------------|
| TXT | @ | `resend._domainkey=...` | DNS only |
| MX | @ | `feedback-smtp.resend.com` (priority 10) | DNS only |

**Example records you'll see:**
```
TXT Record:
Name: @
Content: resend._domainkey.booklib.app=p=MIGfMA0GCSqGSIb3DQEB...

MX Record:
Name: @
Content: feedback-smtp.resend.com
Priority: 10
```

#### 4️⃣ Verify Domain

1. After adding DNS records, go back to Resend
2. Click **Verify DNS Records**
3. Wait 1-2 minutes for DNS propagation
4. Status should change to ✅ **Verified**

#### 5️⃣ Create API Key

1. In Resend, go to **API Keys**
2. Click **Create API Key**
3. Name it: `booklib-backend-production`
4. Select permission: **Sending access**
5. Click **Add**
6. **Copy the API key** - you won't see it again!

#### 6️⃣ Get SMTP Credentials

Resend provides SMTP access:

```
SMTP Host: smtp.resend.com
SMTP Port: 465 (SSL) or 587 (TLS)
SMTP Username: resend
SMTP Password: <your-api-key-from-step-5>
```

#### 7️⃣ Update Railway Environment Variables

Go to your Railway project and update these variables:

```env
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_USERNAME=resend
SMTP_PASSWORD=re_YourApiKey123...
SMTP_FROM_EMAIL=no-reply@booklib.app
SMTP_FROM_NAME=BookLib
```

**Via Railway Dashboard:**
1. Go to https://railway.app/dashboard
2. Select `booklib-backend`
3. Go to **Variables** tab
4. Add/update the variables above
5. Click **Save** (Railway will redeploy)

**Via Railway CLI:**
```bash
cd /Users/brandon/personal/booklib-backend
railway variables set SMTP_HOST="smtp.resend.com"
railway variables set SMTP_PORT="587"
railway variables set SMTP_USERNAME="resend"
railway variables set SMTP_PASSWORD="re_YourApiKey123..."
railway variables set SMTP_FROM_EMAIL="no-reply@booklib.app"
railway variables set SMTP_FROM_NAME="BookLib"
```

#### 8️⃣ Test Your Setup

After Railway redeploys (1-2 minutes):

1. Login to your BookLib app
2. Add a book
3. Lend it to someone with a due date in 2-3 days
4. Check Resend dashboard → **Logs** to see if emails are sent
5. Check your email inbox for the reminder

Or test manually via API:
```bash
curl -X POST https://booklib-backend-production.up.railway.app/api/reminders/test \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🎯 Option 2: SendGrid

### Why SendGrid?
- ✅ **Free tier**: 100 emails/day forever
- ✅ **Reliable**: Used by millions
- ✅ **Good for small projects**

### Quick Setup

#### 1️⃣ Create SendGrid Account
1. Go to https://sendgrid.com
2. Sign up (free account)
3. Verify email and complete setup

#### 2️⃣ Create API Key
1. Go to **Settings** → **API Keys**
2. Click **Create API Key**
3. Name: `booklib-backend`
4. Permission: **Restricted Access** → Enable **Mail Send**
5. Copy the API key

#### 3️⃣ Verify Domain
1. Go to **Settings** → **Sender Authentication**
2. Click **Authenticate Your Domain**
3. Select **Cloudflare** as DNS host
4. Add DNS records to Cloudflare (similar to Resend)
5. Verify

#### 4️⃣ Update Railway Variables
```env
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=SG.your-api-key-here
SMTP_FROM_EMAIL=no-reply@booklib.app
SMTP_FROM_NAME=BookLib
```

---

## 📨 Option 3: Cloudflare Email Routing (Receiving Only)

**Note**: Cloudflare Email Routing is **free** but only for **receiving** emails. You still need a sending service (like Resend).

### Setup Receiving
1. Go to Cloudflare Dashboard → Email → Email Routing
2. Enable Email Routing
3. Add destination address (your personal email)
4. Create routing rules:
   - `no-reply@booklib.app` → your email
   - `*@booklib.app` → your email (catch-all)

### For Sending
- Still use Resend, SendGrid, or Postmark for sending
- This just lets you receive replies (though `no-reply` shouldn't get many!)

---

## 🎯 Option 4: Postmark

### Why Postmark?
- ✅ **Excellent deliverability** (best in class)
- ✅ Free tier: 100 emails/month
- ✅ Great for transactional emails
- ❌ Smaller free tier than Resend

### Quick Setup
1. Sign up at https://postmarkapp.com
2. Create a server
3. Add sender signature or domain
4. Get SMTP credentials

```env
SMTP_HOST=smtp.postmarkapp.com
SMTP_PORT=587
SMTP_USERNAME=<your-server-token>
SMTP_PASSWORD=<your-server-token>
SMTP_FROM_EMAIL=no-reply@booklib.app
SMTP_FROM_NAME=BookLib
```

---

## 🧪 Testing Email Setup

### Test from Your App

1. **Quick Test via Reminders**:
   - Lend a book with due date in 2 days
   - Wait for scheduled reminder (or trigger manually)
   - Check email

2. **Manual Trigger** (if you add this endpoint):
```bash
curl -X POST https://booklib-backend-production.up.railway.app/api/test-email \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email": "your-email@example.com"}'
```

### Check Logs
```bash
# Railway logs
railway logs -s booklib-backend

# Check for:
# ✅ "Email sent successfully to..."
# ❌ "Failed to send email: ..."
```

---

## 🔧 Troubleshooting

### DNS Not Verifying
- **Wait**: DNS can take up to 24 hours (usually 5-10 minutes)
- **Check**: Use https://mxtoolbox.com/SuperTool.aspx to verify DNS records
- **Cloudflare**: Make sure proxy is OFF (DNS only) for email records

### Emails Not Sending
- **Check Railway logs**: `railway logs`
- **Verify credentials**: Double-check SMTP username/password
- **Check spam folder**: First emails might go to spam
- **Test SMTP**: Use a tool like https://www.smtper.net

### Emails Going to Spam
- **SPF/DKIM**: Make sure DNS records are verified
- **Warm up**: First few emails might go to spam, gets better over time
- **Content**: Avoid spam trigger words
- **Test**: Use https://www.mail-tester.com

### Port Issues
If port 587 doesn't work, try:
- Port 465 (SSL)
- Port 2525 (alternative)

---

## 📊 Recommended Choice

For BookLib, I recommend **Resend** because:

1. ✅ **Generous free tier**: 3,000 emails/month is plenty
2. ✅ **Simple setup**: Modern, developer-friendly
3. ✅ **SMTP support**: Works with your existing Go code (no code changes!)
4. ✅ **Good deliverability**: Built by email experts
5. ✅ **Dashboard**: See email logs, opens, bounces

---

## 🎉 Next Steps

1. Choose your email provider (I recommend Resend)
2. Follow the setup steps above
3. Update Railway environment variables
4. Test by sending a reminder
5. Monitor email logs in your provider's dashboard

---

## 📚 Additional Resources

- [Resend Documentation](https://resend.com/docs/introduction)
- [SendGrid Go Library](https://github.com/sendgrid/sendgrid-go)
- [Cloudflare Email Routing](https://developers.cloudflare.com/email-routing/)
- [Email Best Practices](https://postmarkapp.com/guides)

---

## 💡 Pro Tips

### Monitor Email Deliverability
- Check bounce rates in your email provider dashboard
- Monitor spam complaints
- Keep an eye on open rates

### Email Best Practices
- Use clear subject lines
- Include unsubscribe option (future feature)
- Don't send too frequently (your 24hr rate limit is good!)
- Test emails before going live

### Cost Monitoring
- **Resend**: 3,000 emails/month free, then $20/month for 50k
- **SendGrid**: 100 emails/day free, then $15/month for 40k
- **Postmark**: 100 emails/month free, then $15/month for 10k

For a personal book tracking app, the free tiers should be plenty! 📚
