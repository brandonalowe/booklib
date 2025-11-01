# 📚 BookLib Backend

A Go-based REST API backend for managing your personal book library. Track books you own, books you've read, lending history, reading sessions, and get personalized statistics.

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ Features

### 📖 Book Management
- **Add books** by ISBN (auto-fetches metadata from Google Books API)
- **Manual entry** for books without ISBNs
- **Bulk import** from CSV files
- **Duplicate prevention** - won't add the same ISBN twice
- **Search** your library by title, author, or genre
- **Mark books as read**

### 📊 Reading Tracking
- **Reading history** - Track when you started and finished reading each book
- **Multiple readings** - Record re-reading the same book
- **Reading sessions** - See what you're currently reading
- **Statistics** - Books read this month/year, reading trends

### 📤 Lending Management
- **Lend books** to friends with due dates
- **Track borrowed books** - See who has what
- **Overdue notifications** - Know when books should be returned
- **Email reminders** (optional) - Automated reminders for overdue books
- **Most lent books** - See your most popular titles

### 📈 Statistics & Insights
- Total books in library
- Books read this month/year
- Favorite genres
- Currently lent books
- Reading history over time
- Most lent books

### 🔐 User Management
- **User authentication** with JWT tokens
- **Admin dashboard** - User management and system statistics
- **Role-based access** - Admin and regular user roles
- **Secure password hashing** with bcrypt

## 🏗️ Tech Stack

- **Language**: Go 1.24
- **Framework**: Chi (HTTP router)
- **Database**: SQLite with WAL mode
- **Authentication**: JWT tokens + bcrypt
- **APIs**: Google Books API for metadata
- **Scheduling**: Cron jobs for automated tasks

## 🚀 Quick Start

### Prerequisites
- Go 1.24 or higher
- SQLite3

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/brandonalowe/booklib-backend.git
   cd booklib-backend
   ```

2. **Create environment file**
   ```bash
   cp .env.example .env
   ```

3. **Edit `.env` with your settings**
   ```bash
   # Generate a secure secret
   openssl rand -base64 32
   
   # Add to .env
   SESSION_SECRET=your-generated-secret-here
   DATABASE_PATH=./database/books.db
   PORT=8080
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Run the server**
   ```bash
   go run ./cmd/server
   ```

6. **Server starts on http://localhost:8080**
   ```
   Test health check: curl http://localhost:8080/health
   ```

## 📁 Project Structure

```
booklib-backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   └── google.go            # Google Books API integration
│   ├── db/
│   │   └── db.go                # Database initialization
│   ├── handlers/
│   │   ├── authHandler.go       # Authentication endpoints
│   │   ├── bookHandler.go       # Book CRUD operations
│   │   ├── lendingHandler.go    # Lending management
│   │   ├── statsHandler.go      # Statistics
│   │   ├── readingHistoryHandler.go # Reading tracking
│   │   └── adminHandler.go      # Admin functions
│   ├── middleware/
│   │   └── authMiddleware.go    # JWT authentication
│   ├── models/
│   │   ├── book.go              # Book model
│   │   ├── user.go              # User model
│   │   ├── lending.go           # Lending model
│   │   └── reading_history.go   # Reading history model
│   └── services/
│       ├── authService.go       # JWT & password handling
│       ├── emailService.go      # Email notifications
│       └── reminderService.go   # Automated reminders
├── scripts/
│   └── backup.sh                # Database backup script
├── .env.example                 # Environment variables template
├── Dockerfile                   # Docker configuration
├── railway.json                 # Railway.app config
└── README.md
```

## 🔌 API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout
- `GET /api/auth/me` - Get current user

### Books
- `GET /api/books` - List all books
- `GET /api/books/{id}` - Get book details
- `POST /api/books` - Add new book
- `PUT /api/books/{id}` - Update book
- `DELETE /api/books/{id}` - Delete book
- `GET /api/books/search` - Search books
- `GET /api/books/search/{isbn}` - Search by ISBN

### Lending
- `GET /api/lending` - List all lendings
- `POST /api/lending` - Lend a book
- `PUT /api/lending/{id}/return` - Return a book
- `GET /api/lending/overdue` - Get overdue books

### Reading History
- `POST /api/reading-history/start` - Start reading a book
- `PUT /api/reading-history/{id}/finish` - Finish reading
- `GET /api/reading-history/book/{bookId}` - Get reading history for a book
- `GET /api/reading-history/book/{bookId}/active` - Get active reading session

### Statistics
- `GET /api/stats` - Get user statistics

### Admin (Admin role required)
- `GET /api/admin/stats` - System-wide statistics
- `GET /api/admin/users` - List all users
- `GET /api/admin/users/{id}` - Get user details
- `DELETE /api/admin/users/{id}` - Delete user
- `PUT /api/admin/users/{id}/role` - Update user role

### Health
- `GET /health` - Health check endpoint

## 🚢 Deployment

### Railway.app (Recommended - FREE)

**See [RAILWAY_QUICKSTART.md](./RAILWAY_QUICKSTART.md) for detailed instructions.**

Quick deploy:
```bash
npm install -g @railway/cli
railway login
railway init
railway up
```

**Don't forget to:**
1. Add a persistent volume at `/data` (1GB)
2. Set environment variables (`SESSION_SECRET`, `DATABASE_PATH`, etc.)

### Alternative Platforms

See [DEPLOYMENT.md](./DEPLOYMENT.md) for other hosting options:
- Fly.io
- Render.com
- Koyeb
- Oracle Cloud

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_PATH` | SQLite database file path | `./database/books.db` | Yes |
| `PORT` | Server port | `8080` | No |
| `SESSION_SECRET` | Secret for JWT tokens | - | Yes |
| `CORS_ALLOWED_ORIGINS` | Comma-separated allowed origins | `http://localhost:5173` | Yes |
| `REMINDER_CRON_SCHEDULE` | Cron schedule for reminders | `0 9 * * *` | No |
| `RUN_REMINDERS_ON_STARTUP` | Run reminders on startup | `false` | No |
| `SMTP_HOST` | SMTP server host | - | No |
| `SMTP_PORT` | SMTP server port | - | No |
| `SMTP_USERNAME` | SMTP username | - | No |
| `SMTP_PASSWORD` | SMTP password | - | No |
| `SMTP_FROM_EMAIL` | Sender email address | - | No |
| `SMTP_FROM_NAME` | Sender name | `BookLib` | No |

### Email Reminders (Optional)

To enable email reminders for overdue books, configure SMTP settings in `.env`:

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@booklib.com
SMTP_FROM_NAME=BookLib
```

## 🗄️ Database

BookLib uses SQLite with WAL mode for better concurrent access. The database includes:

- **users** - User accounts
- **books** - Book library
- **isbn_cache** - Google Books API response cache
- **lendings** - Book lending records
- **reading_history** - Reading session tracking

### Backups

For production deployments, use the included backup script:

```bash
./scripts/backup.sh
```

Railway.app provides volume snapshots. See deployment docs for details.

## 🧪 Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o booklib ./cmd/server
./booklib
```

### Docker
```bash
docker build -t booklib-backend .
docker run -p 8080:8080 \
  -e SESSION_SECRET="your-secret" \
  -e DATABASE_PATH="/data/booklib.db" \
  -v booklib-data:/data \
  booklib-backend
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 🔗 Related

- **Frontend**: [booklib-frontend](https://github.com/brandonalowe/booklib-frontend) - Vue.js frontend for BookLib

## 📧 Support

For issues, questions, or suggestions, please open an issue on GitHub.

---

Made with ❤️ for book lovers everywhere 📚
