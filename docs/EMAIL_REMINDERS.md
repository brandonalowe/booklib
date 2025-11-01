# Email Reminder System Setup Guide

## Overview

The BookLib email reminder system automatically sends email notifications to users about their lent books:

- **Upcoming Due Reminder**: Sent 3 days before the due date
- **Overdue Reminder**: Sent when a book becomes overdue
- **Rate Limiting**: Only sends one reminder per book every 24 hours to avoid spam

## Prerequisites

You'll need access to an SMTP email server. Common options:

### Option 1: Gmail (Recommended for Testing)

1. Enable 2-Factor Authentication on your Gmail account
2. Generate an App Password:
   - Go to Google Account → Security → 2-Step Verification → App passwords
   - Select "Mail" and your device
   - Copy the generated 16-character password

### Option 2: SendGrid (Recommended for Production)

1. Sign up at [SendGrid](https://sendgrid.com/)
2. Create an API key with "Mail Send" permissions
3. Use the API key as your SMTP password

### Option 3: Other SMTP Providers

- **Mailgun**: smtp.mailgun.org:587
- **Outlook/Office365**: smtp-mail.outlook.com:587
- **AWS SES**: email-smtp.us-east-1.amazonaws.com:587

## Configuration

### 1. Copy Environment File

```bash
cd booklib-backend
cp .env.example .env
```

### 2. Configure SMTP Settings

Edit `.env` and add your SMTP credentials:

```env
# Gmail Example
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@booklib.com
SMTP_FROM_NAME=BookLib

# Cron Schedule (default: 9 AM daily)
REMINDER_CRON_SCHEDULE=0 9 * * *

# Run reminders on startup (useful for testing)
RUN_REMINDERS_ON_STARTUP=true
```

### 3. Load Environment Variables

The server automatically loads `.env` files if you have `github.com/joho/godotenv` installed (already in dependencies).

If not already done, add to `main.go`:

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file
    godotenv.Load()
    
    // ... rest of code
}
```

## Cron Schedule Format

The `REMINDER_CRON_SCHEDULE` uses standard cron syntax:

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

### Examples:

```env
# Every day at 9 AM
REMINDER_CRON_SCHEDULE=0 9 * * *

# Every day at 9 AM and 5 PM
REMINDER_CRON_SCHEDULE=0 9,17 * * *

# Every hour
REMINDER_CRON_SCHEDULE=0 * * * *

# Every 30 minutes (for testing)
REMINDER_CRON_SCHEDULE=*/30 * * * *

# Weekdays only at 9 AM
REMINDER_CRON_SCHEDULE=0 9 * * 1-5
```

## Testing

### Test Email Sending

1. **Set run on startup**:
   ```env
   RUN_REMINDERS_ON_STARTUP=true
   ```

2. **Create a test lending** with a due date 3 days from now or in the past

3. **Start the server**:
   ```bash
   go run cmd/server/main.go
   ```

4. **Check logs** for:
   ```
   Running initial reminder check on startup...
   Starting reminder check...
   Sent X upcoming due reminders
   Sent X overdue reminders
   ```

### Verify Email Configuration

The service will log warnings if SMTP is not configured:

```
Email service not configured, skipping email send
```

If you see this, double-check your `.env` file has valid credentials.

## Email Templates

Three email types are included:

### 1. Upcoming Due Reminder (Individual)
- Sent 3 days before due date
- Purple/indigo color scheme
- Shows book details and days until due
- One email per book

### 2. Overdue Digest Email (Consolidated)
- Sent when books are past due date
- Red color scheme with urgency
- **One email per user** listing ALL their overdue books
- Shows book details with days overdue for each
- Includes summary count
- Much better user experience than multiple individual emails

All templates are mobile-responsive and work in all major email clients.

## Database Schema

The `lending` table includes a `last_reminder_sent` column:

```sql
CREATE TABLE lending (
    ...
    last_reminder_sent DATETIME,
    ...
);
```

This prevents spam by ensuring reminders are only sent once per 24-hour period.

## Reminder Logic

### Upcoming Due Reminders

```sql
-- Finds books due in exactly 3 days
-- that haven't been reminded in the last 24 hours
WHERE l.returned_at IS NULL
  AND l.due_date IS NOT NULL
  AND DATE(l.due_date) = DATE('now', '+3 days')
  AND (l.last_reminder_sent IS NULL OR l.last_reminder_sent < datetime('now', '-24 hours'))
```

### Overdue Reminders

```sql
-- Finds books past their due date
-- that haven't been reminded in the last 24 hours
WHERE l.returned_at IS NULL
  AND l.due_date IS NOT NULL
  AND DATE(l.due_date) < DATE('now')
  AND (l.last_reminder_sent IS NULL OR l.last_reminder_sent < datetime('now', '-24 hours'))
```

## Troubleshooting

### Emails Not Sending

1. **Check SMTP credentials** in `.env`
2. **Check firewall** - ensure port 587 is not blocked
3. **Gmail users** - verify App Password (not regular password)
4. **Check logs** for specific error messages

### Gmail "Less Secure Apps" Error

Gmail no longer supports "less secure apps". You MUST use an App Password with 2FA enabled.

### SendGrid/Mailgun Issues

Verify your API key has the correct permissions and your account is verified.

### Testing Without SMTP

If you don't want to configure SMTP yet, the system will log reminders without sending:

```
Email service not configured, skipping email send
```

The database will still be updated with `last_reminder_sent` times.

## Production Recommendations

1. **Use a dedicated email service** (SendGrid, Mailgun, AWS SES)
2. **Set up SPF/DKIM** records to improve deliverability
3. **Monitor** email send rates and bounces
4. **Add unsubscribe** functionality if needed
5. **Use a no-reply email** address for sending
6. **Schedule during business hours** (9 AM is good)
7. **Consider timezone** of your users

## Future Enhancements

- Add user preferences for reminder frequency
- Include a "View in Browser" link
- Add weekly digest for multiple overdue books
- Track email open rates
- Add SMS/push notification options
- Allow users to opt-out of reminders

