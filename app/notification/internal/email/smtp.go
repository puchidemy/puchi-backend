package email

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

// NewFromEnv creates a Sender from SMTP_* environment variables.
func NewFromEnv(logger *slog.Logger) *Sender {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	return &Sender{
		host:       host,
		port:       port,
		username:   os.Getenv("SMTP_USERNAME"),
		password:   os.Getenv("SMTP_PASSWORD"),
		fromEmail:  fromEmail,
		fromName:   os.Getenv("SMTP_FROM_NAME"),
		configured: host != "" && port != "" && fromEmail != "",
		logger:     logger,
	}
}

// IsConfigured returns true if enough SMTP config is present to send.
func (s *Sender) IsConfigured() bool {
	return s.configured
}

// Send sends an email via SMTP. If not configured, logs instead of sending.
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

	s.logger.Info("email sent", "to", to, "subject", subject, "from", s.fromEmail)
	return nil
}

func buildMIMEMessage(from, to, subject, htmlBody, plainBody string) string {
	boundary := "boundary-puchi-email"

	msg := fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	msg += "\r\n"

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

	msg += fmt.Sprintf("--%s\r\n", boundary)
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "Content-Transfer-Encoding: quoted-printable\r\n"
	msg += "\r\n"
	msg += htmlBody + "\r\n"
	msg += fmt.Sprintf("--%s--\r\n", boundary)

	return msg
}
