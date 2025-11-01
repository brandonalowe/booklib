package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"time"
)

type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type EmailData struct {
	UserEmail   string
	BookTitle   string
	BookAuthor  string
	LentTo      string
	DueDate     time.Time
	LentAt      time.Time
	DaysOverdue int
}

type OverdueBook struct {
	BookTitle   string
	BookAuthor  string
	LentTo      string
	DueDate     time.Time
	DaysOverdue int
}

type OverdueDigestData struct {
	UserEmail    string
	OverdueBooks []OverdueBook
	TotalOverdue int
}

func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("SMTP_FROM_EMAIL", "noreply@booklib.com"),
		FromName:     getEnv("SMTP_FROM_NAME", "BookLib"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (e *EmailService) IsConfigured() bool {
	return e.SMTPUsername != "" && e.SMTPPassword != ""
}

func (e *EmailService) SendUpcomingDueReminder(data EmailData) error {
	subject := "Reminder: Book due soon"
	body := e.renderUpcomingDueTemplate(data)
	return e.sendEmail(data.UserEmail, subject, body)
}

func (e *EmailService) SendOverdueReminder(data EmailData) error {
	subject := "Reminder: Book is overdue"
	body := e.renderOverdueTemplate(data)
	return e.sendEmail(data.UserEmail, subject, body)
}

func (e *EmailService) SendOverdueDigest(data OverdueDigestData) error {
	subject := fmt.Sprintf("Reminder: You have %d overdue book(s)", data.TotalOverdue)
	body := e.renderOverdueDigestTemplate(data)
	return e.sendEmail(data.UserEmail, subject, body)
}

func (e *EmailService) sendEmail(to, subject, htmlBody string) error {
	if !e.IsConfigured() {
		log.Println("Email service not configured, skipping email send")
		return fmt.Errorf("email service not configured")
	}

	from := fmt.Sprintf("%s <%s>", e.FromName, e.FromEmail)

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	auth := smtp.PlainAuth("", e.SMTPUsername, e.SMTPPassword, e.SMTPHost)
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)

	err := smtp.SendMail(addr, auth, e.FromEmail, []string{to}, []byte(message))
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("Successfully sent email to %s", to)
	return nil
}

func (e *EmailService) renderUpcomingDueTemplate(data EmailData) string {
	daysUntilDue := int(time.Until(data.DueDate).Hours() / 24)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background-color: #f9fafb; padding: 30px; border: 1px solid #e5e7eb; border-radius: 0 0 8px 8px; }
        .book-info { background-color: white; padding: 20px; margin: 20px 0; border-radius: 8px; border-left: 4px solid #4F46E5; }
        .info-row { margin: 10px 0; }
        .label { font-weight: bold; color: #6b7280; }
        .value { color: #111827; }
        .footer { text-align: center; margin-top: 30px; color: #6b7280; font-size: 12px; }
        .warning { background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📚 BookLib Reminder</h1>
    </div>
    <div class="content">
        <h2>Book Due Soon</h2>
        <p>Hi there,</p>
        <p>This is a friendly reminder that a book you lent out is due in <strong>{{.DaysUntilDue}} days</strong>.</p>
        
        <div class="book-info">
            <h3>📖 Book Details</h3>
            <div class="info-row">
                <span class="label">Title:</span> 
                <span class="value">{{.BookTitle}}</span>
            </div>
            <div class="info-row">
                <span class="label">Author:</span> 
                <span class="value">{{.BookAuthor}}</span>
            </div>
            <div class="info-row">
                <span class="label">Lent to:</span> 
                <span class="value">{{.LentTo}}</span>
            </div>
            <div class="info-row">
                <span class="label">Due date:</span> 
                <span class="value">{{.DueDateFormatted}}</span>
            </div>
        </div>

        <div class="warning">
            <strong>⏰ Action Needed:</strong> You may want to reach out to {{.LentTo}} to remind them about the upcoming due date.
        </div>

        <p>Thank you for using BookLib!</p>
    </div>
    <div class="footer">
        <p>This is an automated message from BookLib. Please do not reply to this email.</p>
    </div>
</body>
</html>
`

	t, err := template.New("upcomingDue").Parse(tmpl)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return ""
	}

	templateData := struct {
		BookTitle        string
		BookAuthor       string
		LentTo           string
		DueDateFormatted string
		DaysUntilDue     int
	}{
		BookTitle:        data.BookTitle,
		BookAuthor:       data.BookAuthor,
		LentTo:           data.LentTo,
		DueDateFormatted: data.DueDate.Format("Monday, January 2, 2006"),
		DaysUntilDue:     daysUntilDue,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, templateData); err != nil {
		log.Printf("Error executing template: %v", err)
		return ""
	}

	return buf.String()
}

func (e *EmailService) renderOverdueTemplate(data EmailData) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #DC2626; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background-color: #f9fafb; padding: 30px; border: 1px solid #e5e7eb; border-radius: 0 0 8px 8px; }
        .book-info { background-color: white; padding: 20px; margin: 20px 0; border-radius: 8px; border-left: 4px solid #DC2626; }
        .info-row { margin: 10px 0; }
        .label { font-weight: bold; color: #6b7280; }
        .value { color: #111827; }
        .footer { text-align: center; margin-top: 30px; color: #6b7280; font-size: 12px; }
        .alert { background-color: #fee2e2; border-left: 4px solid #DC2626; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .overdue-badge { background-color: #DC2626; color: white; padding: 5px 10px; border-radius: 4px; font-weight: bold; display: inline-block; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>⚠️ BookLib Overdue Notice</h1>
    </div>
    <div class="content">
        <h2>Book is Overdue</h2>
        <p>Hi there,</p>
        <p>A book you lent out is now overdue.</p>
        
        <div class="overdue-badge">
            OVERDUE BY {{.DaysOverdue}} DAYS
        </div>

        <div class="book-info">
            <h3>📖 Book Details</h3>
            <div class="info-row">
                <span class="label">Title:</span> 
                <span class="value">{{.BookTitle}}</span>
            </div>
            <div class="info-row">
                <span class="label">Author:</span> 
                <span class="value">{{.BookAuthor}}</span>
            </div>
            <div class="info-row">
                <span class="label">Lent to:</span> 
                <span class="value">{{.LentTo}}</span>
            </div>
            <div class="info-row">
                <span class="label">Was due:</span> 
                <span class="value">{{.DueDateFormatted}}</span>
            </div>
        </div>

        <div class="alert">
            <strong>🔔 Please Follow Up:</strong> We recommend contacting {{.LentTo}} to request the return of this book.
        </div>

        <p>Thank you for using BookLib!</p>
    </div>
    <div class="footer">
        <p>This is an automated message from BookLib. Please do not reply to this email.</p>
    </div>
</body>
</html>
`

	t, err := template.New("overdue").Parse(tmpl)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return ""
	}

	templateData := struct {
		BookTitle        string
		BookAuthor       string
		LentTo           string
		DueDateFormatted string
		DaysOverdue      int
	}{
		BookTitle:        data.BookTitle,
		BookAuthor:       data.BookAuthor,
		LentTo:           data.LentTo,
		DueDateFormatted: data.DueDate.Format("Monday, January 2, 2006"),
		DaysOverdue:      data.DaysOverdue,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, templateData); err != nil {
		log.Printf("Error executing template: %v", err)
		return ""
	}

	return buf.String()
}

func (e *EmailService) renderOverdueDigestTemplate(data OverdueDigestData) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #DC2626; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background-color: #f9fafb; padding: 30px; border: 1px solid #e5e7eb; border-radius: 0 0 8px 8px; }
        .book-item { background-color: white; padding: 15px; margin: 15px 0; border-radius: 8px; border-left: 4px solid #DC2626; }
        .book-title { font-weight: bold; font-size: 16px; color: #111827; margin-bottom: 8px; }
        .book-detail { font-size: 14px; color: #6b7280; margin: 4px 0; }
        .overdue-badge { background-color: #DC2626; color: white; padding: 4px 8px; border-radius: 4px; font-weight: bold; font-size: 12px; display: inline-block; }
        .footer { text-align: center; margin-top: 30px; color: #6b7280; font-size: 12px; }
        .summary { background-color: #fee2e2; border-left: 4px solid #DC2626; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .summary-count { font-size: 24px; font-weight: bold; color: #DC2626; }
    </style>
</head>
<body>
    <div class="header">
        <h1>⚠️ BookLib Overdue Notice</h1>
    </div>
    <div class="content">
        <h2>You Have Overdue Books</h2>
        <p>Hi there,</p>
        <p>You have <strong>{{.TotalOverdue}} book(s)</strong> that are currently overdue. Please follow up with the borrowers to request their return.</p>
        
        <div class="summary">
            <div class="summary-count">{{.TotalOverdue}} Overdue Book{{if ne .TotalOverdue 1}}s{{end}}</div>
        </div>

        <h3>📚 Overdue Books:</h3>
        
        {{range .OverdueBooks}}
        <div class="book-item">
            <div class="book-title">📖 {{.BookTitle}}</div>
            <div class="book-detail"><strong>Author:</strong> {{.BookAuthor}}</div>
            <div class="book-detail"><strong>Lent to:</strong> {{.LentTo}}</div>
            <div class="book-detail"><strong>Was due:</strong> {{.DueDateFormatted}}</div>
            <div class="book-detail">
                <span class="overdue-badge">OVERDUE BY {{.DaysOverdue}} DAY{{if ne .DaysOverdue 1}}S{{end}}</span>
            </div>
        </div>
        {{end}}

        <div style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px;">
            <strong>🔔 Action Needed:</strong> We recommend reaching out to these borrowers to request the return of your books.
        </div>

        <p>Thank you for using BookLib!</p>
    </div>
    <div class="footer">
        <p>This is an automated message from BookLib. Please do not reply to this email.</p>
    </div>
</body>
</html>
`

	t, err := template.New("overdueDigest").Parse(tmpl)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return ""
	}

	// Format the overdue books with dates
	type FormattedBook struct {
		BookTitle        string
		BookAuthor       string
		LentTo           string
		DueDateFormatted string
		DaysOverdue      int
	}

	var formattedBooks []FormattedBook
	for _, book := range data.OverdueBooks {
		formattedBooks = append(formattedBooks, FormattedBook{
			BookTitle:        book.BookTitle,
			BookAuthor:       book.BookAuthor,
			LentTo:           book.LentTo,
			DueDateFormatted: book.DueDate.Format("Monday, January 2, 2006"),
			DaysOverdue:      book.DaysOverdue,
		})
	}

	templateData := struct {
		TotalOverdue int
		OverdueBooks []FormattedBook
	}{
		TotalOverdue: data.TotalOverdue,
		OverdueBooks: formattedBooks,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, templateData); err != nil {
		log.Printf("Error executing template: %v", err)
		return ""
	}

	return buf.String()
}
