# 📚 BookLib Backend

Go REST API for managing your personal book library. Track books, reading history, lending, and get statistics.

## ✨ Features

- 📖 **Book Management** - Add books by ISBN (auto-fetch metadata) or manually
- 📊 **Reading Tracking** - Track reading sessions, history, and re-reads
- 📤 **Lending System** - Lend books to friends with due dates and overdue tracking
- 📧 **Email Reminders** - Optional automated overdue notifications
- 📈 **Statistics** - Reading trends, favorite genres, lending insights
- 🔐 **Authentication** - JWT-based auth with admin panel

## 🏗️ Tech Stack

Go 1.24 • Chi Router • SQLite • JWT • Google Books API • Cron

## 🚀 Quick Start

**Prerequisites**: Go 1.24+

```bash
# Clone and setup
git clone https://github.com/brandonalowe/booklib-backend.git
cd booklib-backend
cp .env.example .env

# Generate secret
openssl rand -base64 32
# Add to .env: SESSION_SECRET=<generated-secret>

# Run
go mod download
go run ./cmd/server
# Server: http://localhost:8080
```

## 🚢 Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for Railway deployment (10 minutes, free tier).

## 📁 Project Structure

```
cmd/server/        # Entry point
internal/
  ├── api/         # Google Books API
  ├── handlers/    # HTTP handlers
  ├── middleware/  # Auth middleware
  ├── models/      # Data models
  └── services/    # Business logic
scripts/           # Utilities
```

## 🔌 Key API Endpoints

### Auth
- `POST /api/auth/register` - Register
- `POST /api/auth/login` - Login
- `GET /api/auth/me` - Current user

### Books
- `GET /api/books` - List books
- `POST /api/books` - Add book
- `GET /api/books/search/{isbn}` - ISBN lookup

### Lending
- `POST /api/lending` - Lend book
- `PUT /api/lending/{id}/return` - Return book
- `GET /api/lending/overdue` - Overdue books

### Reading
- `POST /api/reading-history/start` - Start reading
- `PUT /api/reading-history/{id}/finish` - Finish reading

### Stats
- `GET /api/stats` - User statistics

## ⚙️ Configuration

### Required Environment Variables

```bash
SESSION_SECRET=<random-secret>     # Generate: openssl rand -base64 32
DATABASE_PATH=/data/booklib.db
PORT=8080
CORS_ALLOWED_ORIGINS=https://your-frontend.pages.dev
```

### Optional: Email Reminders

See [docs/EMAIL_SETUP_CUSTOM_DOMAIN.md](docs/EMAIL_SETUP_CUSTOM_DOMAIN.md)

```bash
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_USERNAME=resend
SMTP_PASSWORD=your-api-key
SMTP_FROM_EMAIL=no-reply@yourdomain.com
```

## 🗄️ Database

SQLite with WAL mode. Includes: users, books, lendings, reading_history, isbn_cache.

**Backup**: `./scripts/backup.sh` or use Railway volume snapshots.

## 🧪 Development

```bash
go test ./...                    # Run tests
go build -o booklib ./cmd/server # Build
docker build -t booklib .        # Docker build
```

## 🔗 Related

**Frontend**: [booklib-frontend](https://github.com/brandonalowe/booklib-frontend)

---

Made with ❤️ for book lovers 📚
