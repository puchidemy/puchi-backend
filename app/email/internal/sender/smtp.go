package sender

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
)

// Sender sends emails via SMTP.
type Sender struct {
	host       string
	port       string
	username   string
	password   string
	fromEmail  string
	fromName   string
	configured bool
	logger     *slog.Logger
}

// New creates a new Sender with the given SMTP configuration.
func New(host, port, username, password, fromEmail, fromName string, logger *slog.Logger) *Sender {
	return &Sender{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		fromEmail:  fromEmail,
		fromName:   fromName,
		configured: host != "" && port != "" && fromEmail != "",
		logger:     logger,
	}
}

// NewFromEnv creates a Sender from environment variables.
func NewFromEnv(logger *slog.Logger) *Sender {
	return &Sender{
		host:       os.Getenv("SMTP_HOST"),
		port:       os.Getenv("SMTP_PORT"),
		username:   os.Getenv("SMTP_USERNAME"),
		password:   os.Getenv("SMTP_PASSWORD"),
		fromEmail:  os.Getenv("SMTP_FROM_EMAIL"),
		fromName:   os.Getenv("SMTP_FROM_NAME"),
		configured: os.Getenv("SMTP_HOST") != "" && os.Getenv("SMTP_PORT") != "" && os.Getenv("SMTP_FROM_EMAIL") != "",
		logger:     logger,
	}
}

// IsConfigured returns true if the SMTP client has enough config to send emails.
func (s *Sender) IsConfigured() bool {
	return s.configured
}

// Send sends an email via SMTP. If the sender is not configured, it logs the email instead.
func (s *Sender) Send(to, subject, body, plainBody string) error {
	if !s.configured {
		s.logger.Warn("SMTP not configured — logging email instead of sending",
			"to", to,
			"subject", subject,
			"plain_body", plainBody,
		)
		return nil
	}

	fromName := s.fromName
	if fromName == "" {
		fromName = "Puchi"
	}

	from := fmt.Sprintf("%s <%s>", fromName, s.fromEmail)
	msg := buildMIMEMessage(from, to, subject, body, plainBody)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	if err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("send email via SMTP: %w", err)
	}

	s.logger.Info("email sent",
		"to", to,
		"subject", subject,
		"from", s.fromEmail,
	)
	return nil
}

// buildMIMEMessage constructs a MIME multipart email message with both HTML and plain text alternatives.
func buildMIMEMessage(from, to, subject, htmlBody, plainBody string) string {
	boundary := "boundary-puchi-email"

	msg := fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	msg += "\r\n"

	// Plain text part
	msg += fmt.Sprintf("--%s\r\n", boundary)
	msg += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	msg += "Content-Transfer-Encoding: quoted-printable\r\n"
	msg += "\r\n"
	if plainBody != "" {
		msg += plainBody + "\r\n"
	} else {
		msg += "Please view this email in an HTML-compatible email client.\r\n"
	}
	msg += "\r\n"

	// HTML part
	msg += fmt.Sprintf("--%s\r\n", boundary)
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "Content-Transfer-Encoding: quoted-printable\r\n"
	msg += "\r\n"
	msg += htmlBody + "\r\n"

	msg += fmt.Sprintf("--%s--\r\n", boundary)

	return msg
}
