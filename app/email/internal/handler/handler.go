package handler

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/puchidemy/puchi-backend/app/email/internal/sender"
)

// EmailMessage represents the payload of an email.send NATS event.
type EmailMessage struct {
	To        string         `json:"to"`
	Subject   string         `json:"subject"`
	Body      string         `json:"body"`
	PlainBody string         `json:"plain_body"`
	Template  string         `json:"template"`
	Data      map[string]any `json:"data"`
}

// Handler processes email.send NATS messages.
type Handler struct {
	sender *sender.Sender
	logger *slog.Logger
}

// New creates a new Handler.
func New(sender *sender.Sender, logger *slog.Logger) *Handler {
	return &Handler{
		sender: sender,
		logger: logger,
	}
}

// HandleMessage handles an incoming NATS message on the email.send topic.
func (h *Handler) HandleMessage(msg *nats.Msg) {
	var email EmailMessage
	if err := json.Unmarshal(msg.Data, &email); err != nil {
		h.logger.Error("parse email message", "error", err)
		return
	}

	// Ensure required fields
	email.To = strings.TrimSpace(email.To)
	if email.To == "" {
		h.logger.Error("email missing 'to' field")
		return
	}

	// Build subject and body from template if not already set
	if email.Subject == "" || email.Body == "" {
		h.buildFromTemplate(&email)
	}

	if err := h.sender.Send(email.To, email.Subject, email.Body, email.PlainBody); err != nil {
		h.logger.Error("send email", "error", err, "to", email.To, "template", email.Template)
		// Don't nats.Nak() — the message will not be re-delivered for now
		return
	}

	h.logger.Info("email processed",
		"to", email.To,
		"template", email.Template,
		"subject", email.Subject,
	)
}

// buildFromTemplate populates subject and body from the template name and data.
func (h *Handler) buildFromTemplate(email *EmailMessage) {
	switch email.Template {
	case "magic-link":
		h.buildMagicLink(email)
	case "password-reset":
		h.buildPasswordReset(email)
	case "email-verify":
		h.buildEmailVerify(email)
	default:
		// If no template matched and body/subject are empty, log a warning
		if email.Subject == "" || email.Body == "" {
			h.logger.Warn("unknown email template and no subject/body provided",
				"template", email.Template)
		}
	}
}

func (h *Handler) buildMagicLink(email *EmailMessage) {
	link, _ := email.Data["link"].(string)
	userName, _ := email.Data["user_name"].(string)
	if userName == "" {
		userName = "there"
	}

	if email.Subject == "" {
		email.Subject = "Your magic link to sign in to Puchi"
	}

	if email.Body == "" {
		email.Body = magicLinkHTML(link, userName)
	}

	if email.PlainBody == "" {
		email.PlainBody = fmtPlainBody(
			"Sign in to Puchi",
			"Click the link below to sign in:",
			link,
			"This link will expire in 15 minutes.",
		)
	}
}

func (h *Handler) buildPasswordReset(email *EmailMessage) {
	link, _ := email.Data["link"].(string)
	userName, _ := email.Data["user_name"].(string)
	if userName == "" {
		userName = "there"
	}

	if email.Subject == "" {
		email.Subject = "Reset your Puchi password"
	}

	if email.Body == "" {
		email.Body = passwordResetHTML(link, userName)
	}

	if email.PlainBody == "" {
		email.PlainBody = fmtPlainBody(
			"Reset your Puchi password",
			"Click the link below to reset your password:",
			link,
			"This link will expire in 15 minutes. If you didn't request this, ignore this email.",
		)
	}
}

func (h *Handler) buildEmailVerify(email *EmailMessage) {
	code, _ := email.Data["code"].(string)
	userName := ""
	if email.Data != nil {
		if emailVal, ok := email.Data["email"].(string); ok {
			userName = strings.Split(emailVal, "@")[0]
		}
	}
	if userName == "" {
		userName = "there"
	}

	if email.Subject == "" {
		email.Subject = "Verify your email address for Puchi"
	}

	if email.Body == "" {
		email.Body = emailVerifyHTML(code, userName)
	}

	if email.PlainBody == "" {
		email.PlainBody = "Verify your email address for Puchi\n\n" +
			"Your verification code is: " + code + "\n\n" +
			"This code will expire in 15 minutes."
	}
}

// --- HTML templates ---

func magicLinkHTML(link, userName string) string {
	return `<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f6f8fa; margin: 0; padding: 0;">
<table width="100%" cellpadding="0" cellspacing="0" style="background: #f6f8fa; padding: 40px 0;">
<tr><td align="center">
<table width="480" cellpadding="0" cellspacing="0" style="background: #ffffff; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.08);">
<tr><td style="padding: 40px 32px 24px;" align="center">
<img src="https://puchi.io.vn/logo.png" alt="Puchi" width="48" height="48" style="border-radius: 12px;" />
<h1 style="font-size: 20px; color: #1a1a2e; margin: 16px 0 4px;">Hi ` + userName + `!</h1>
<p style="color: #64748b; font-size: 14px; margin: 0;">Click the button below to sign in to your account.</p>
<a href="` + link + `" style="display: inline-block; margin: 24px 0; padding: 12px 32px; background: #6366f1; color: #ffffff; text-decoration: none; border-radius: 8px; font-size: 15px; font-weight: 600;">Sign in to Puchi</a>
<p style="color: #94a3b8; font-size: 12px;">This link expires in 15 minutes. If you didn't request this, you can safely ignore this email.</p>
</td></tr>
<tr><td style="padding: 16px 32px; background: #f8fafc; border-top: 1px solid #e2e8f0; border-radius: 0 0 12px 12px;" align="center">
<p style="color: #94a3b8; font-size: 11px; margin: 0;">&copy; 2026 Puchi. All rights reserved.</p>
</td></tr>
</table>
</td></tr>
</table>
</body></html>`
}

func passwordResetHTML(link, userName string) string {
	return `<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f6f8fa; margin: 0; padding: 0;">
<table width="100%" cellpadding="0" cellspacing="0" style="background: #f6f8fa; padding: 40px 0;">
<tr><td align="center">
<table width="480" cellpadding="0" cellspacing="0" style="background: #ffffff; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.08);">
<tr><td style="padding: 40px 32px 24px;" align="center">
<img src="https://puchi.io.vn/logo.png" alt="Puchi" width="48" height="48" style="border-radius: 12px;" />
<h1 style="font-size: 20px; color: #1a1a2e; margin: 16px 0 4px;">Hi ` + userName + `!</h1>
<p style="color: #64748b; font-size: 14px; margin: 0;">We received a request to reset your password.</p>
<a href="` + link + `" style="display: inline-block; margin: 24px 0; padding: 12px 32px; background: #6366f1; color: #ffffff; text-decoration: none; border-radius: 8px; font-size: 15px; font-weight: 600;">Reset Password</a>
<p style="color: #94a3b8; font-size: 12px;">This link expires in 15 minutes. If you didn't request this, you can safely ignore this email.</p>
</td></tr>
<tr><td style="padding: 16px 32px; background: #f8fafc; border-top: 1px solid #e2e8f0; border-radius: 0 0 12px 12px;" align="center">
<p style="color: #94a3b8; font-size: 11px; margin: 0;">&copy; 2026 Puchi. All rights reserved.</p>
</td></tr>
</table>
</td></tr>
</table>
</body></html>`
}

func emailVerifyHTML(code, userName string) string {
	return `<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f6f8fa; margin: 0; padding: 0;">
<table width="100%" cellpadding="0" cellspacing="0" style="background: #f6f8fa; padding: 40px 0;">
<tr><td align="center">
<table width="480" cellpadding="0" cellspacing="0" style="background: #ffffff; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.08);">
<tr><td style="padding: 40px 32px 24px;" align="center">
<img src="https://puchi.io.vn/logo.png" alt="Puchi" width="48" height="48" style="border-radius: 12px;" />
<h1 style="font-size: 20px; color: #1a1a2e; margin: 16px 0 4px;">Hi ` + userName + `!</h1>
<p style="color: #64748b; font-size: 14px; margin: 0;">Enter this code to verify your email address:</p>
<div style="margin: 24px 0; padding: 16px 32px; background: #f0f0ff; border-radius: 8px; border: 1px solid #e0e0ff; font-size: 32px; font-weight: 700; letter-spacing: 8px; color: #6366f1; text-align: center;">` + code + `</div>
<p style="color: #94a3b8; font-size: 12px;">This code expires in 15 minutes. If you didn't create an account, you can safely ignore this email.</p>
</td></tr>
<tr><td style="padding: 16px 32px; background: #f8fafc; border-top: 1px solid #e2e8f0; border-radius: 0 0 12px 12px;" align="center">
<p style="color: #94a3b8; font-size: 11px; margin: 0;">&copy; 2026 Puchi. All rights reserved.</p>
</td></tr>
</table>
</td></tr>
</table>
</body></html>`
}

func fmtPlainBody(title, instruction, link, footer string) string {
	return title + "\n\n" + instruction + "\n" + link + "\n\n" + footer
}
