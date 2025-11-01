# Quick Start: Email Reminders

## Setup in 5 Minutes

### 1. Install Dependencies
```bash
cd booklib-backend
go mod tidy
```

### 2. Create .env File
```bash
cp .env.example .env
```

### 3. Configure Gmail (Easiest for Testing)

**Get an App Password:**
1. Go to https://myaccount.google.com/security
2. Enable "2-Step Verification"
3. Go to "App passwords"
4. Generate password for "Mail" + your device
5. Copy the 16-character password

**Edit .env:**
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=xxxx-xxxx-xxxx-xxxx
SMTP_FROM_EMAIL=your-email@gmail.com
SMTP_FROM_NAME=BookLib

# Test immediately on startup
RUN_REMINDERS_ON_STARTUP=true

# Run every 30 minutes for testing (change to "0 9 * * *" for production)
REMINDER_CRON_SCHEDULE=*/30 * * * *
```

### 4. Start the Server
```bash
go run cmd/server/main.go
```

You should see:
```
No .env file found, using environment variables or defaults
Database initialized successfully
Reminder cron job scheduled: */30 * * * *
Running initial reminder check on startup...
Server starting on :8080
```

### 5. Test the Reminders

**Option A: Create a test lending with upcoming due date**
1. Login to your app
2. Lend a book with due date = 3 days from now
3. Wait for the cron job to run (or restart server if RUN_REMINDERS_ON_STARTUP=true)
4. Check your email

**Option B: Create an overdue lending**
1. Lend a book with due date = yesterday
2. Restart server (or wait for cron)
3. Check your email for overdue notice

### 6. Check Logs

Watch for:
```
Starting reminder check...
Successfully sent email to user@example.com
Sent 1 upcoming due reminders
Sent 1 overdue reminders
Reminder check completed
```

## Production Setup

For production, change your `.env`:

```env
# Daily at 9 AM
REMINDER_CRON_SCHEDULE=0 9 * * *

# Don't run on startup
RUN_REMINDERS_ON_STARTUP=false

# Use a proper email service (SendGrid recommended)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME=Your BookLib
```

## Troubleshooting

**"Email service not configured"**
- Check your SMTP credentials in .env
- Make sure .env is in the booklib-backend directory

**Gmail "Invalid credentials"**
- Use App Password, not your regular password
- Ensure 2FA is enabled

**No emails received**
- Check spam folder
- Check server logs for errors
- Verify email address is correct in database
- Test with `RUN_REMINDERS_ON_STARTUP=true`

**Cron not running**
- Check logs for "Reminder cron job scheduled"
- Verify cron schedule format
- For testing, use `*/5 * * * *` (every 5 minutes)

## What Gets Sent

### Upcoming Due Reminder (3 days before)
- Beautiful HTML email with purple theme
- Shows book title, author, borrower name
- Shows due date
- Only sent once per 24 hours
- **One email per book** (individual reminders)

### Overdue Digest Email
- Red-themed urgent email
- **One consolidated email per user** listing ALL their overdue books
- Shows each book with:
  - Title and author
  - Borrower name
  - Original due date
  - Days overdue count
- Includes summary count at the top
- Only sent once per 24 hours (all books tracked together)
- Much better UX than receiving multiple individual emails!

## Next Steps

- See `docs/EMAIL_REMINDERS.md` for detailed documentation
- Configure your production email service
- Set up SPF/DKIM records for better deliverability
- Consider adding user preferences for reminder frequency

