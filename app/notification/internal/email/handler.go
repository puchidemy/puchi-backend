package email

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"
)

// Message is the payload of an email.send NATS event.
type Message struct {
	To        string         `json:"to"`
	Subject   string         `json:"subject"`
	Body      string         `json:"body"`
	PlainBody string         `json:"plain_body"`
	Template  string         `json:"template"`
	Data      map[string]any `json:"data"`
}

// Handler processes email.send NATS messages.
type Handler struct {
	sender *Sender
	logger *slog.Logger
}

// NewHandler creates a new email Handler.
func NewHandler(sender *Sender, logger *slog.Logger) *Handler {
	return &Handler{sender: sender, logger: logger}
}

// HandleMessage handles an incoming NATS message on email.send.
func (h *Handler) HandleMessage(msg *nats.Msg) {
	var email Message
	if err := json.Unmarshal(msg.Data, &email); err != nil {
		h.logger.Error("parse email message", "error", err)
		return
	}

	email.To = strings.TrimSpace(email.To)
	if email.To == "" {
		h.logger.Error("email missing 'to' field")
		return
	}

	if email.Subject == "" || email.Body == "" {
		h.buildFromTemplate(&email)
	}

	if err := h.sender.Send(email.To, email.Subject, email.Body, email.PlainBody); err != nil {
		h.logger.Error("send email", "error", err, "to", email.To, "template", email.Template)
		return
	}

	h.logger.Info("email processed",
		"to", email.To,
		"template", email.Template,
		"subject", email.Subject,
	)
}

func (h *Handler) buildFromTemplate(email *Message) {
	switch email.Template {
	case "magic-link":
		h.buildMagicLink(email)
	case "password-reset":
		h.buildPasswordReset(email)
	case "email-verify":
		h.buildEmailVerify(email)
	default:
		if email.Subject == "" || email.Body == "" {
			h.logger.Warn("unknown email template and no subject/body provided",
				"template", email.Template)
		}
	}
}

func dataString(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	v, _ := data[key].(string)
	return v
}

func (h *Handler) buildMagicLink(email *Message) {
	link := dataString(email.Data, "link")
	userName := dataString(email.Data, "user_name")
	if userName == "" {
		userName = "there"
	}
	if email.Subject == "" {
		email.Subject = "Your magic link to sign in to Puchi"
	}
	if email.Body == "" {
		email.Body = actionLinkHTML("Sign in to Puchi", userName, "Click the button below to sign in to your account.", "Sign in to Puchi", link)
	}
	if email.PlainBody == "" {
		email.PlainBody = fmtPlainBody("Sign in to Puchi", "Click the link below to sign in:", link, "This link will expire in 15 minutes.")
	}
}

func (h *Handler) buildPasswordReset(email *Message) {
	link := dataString(email.Data, "link")
	userName := dataString(email.Data, "user_name")
	if userName == "" {
		userName = "there"
	}
	if email.Subject == "" {
		email.Subject = "Reset your Puchi password"
	}
	if email.Body == "" {
		email.Body = actionLinkHTML("Reset your password", userName, "We received a request to reset your password.", "Reset Password", link)
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

func (h *Handler) buildEmailVerify(email *Message) {
	link := dataString(email.Data, "link")
	code := dataString(email.Data, "code")
	userName := dataString(email.Data, "user_name")
	if userName == "" {
		if em := dataString(email.Data, "email"); em != "" {
			userName = strings.Split(em, "@")[0]
		}
	}
	if userName == "" {
		userName = "there"
	}

	if email.Subject == "" {
		email.Subject = "Verify your email address for Puchi"
	}

	// Auth publishes link+token (Limen); prefer link CTA. Fall back to code if present.
	if email.Body == "" {
		if link != "" {
			email.Body = actionLinkHTML("Verify your email", userName, "Click the button below to verify your email address.", "Verify Email", link)
		} else {
			email.Body = emailVerifyCodeHTML(code, userName)
		}
	}

	if email.PlainBody == "" {
		if link != "" {
			email.PlainBody = fmtPlainBody(
				"Verify your email address for Puchi",
				"Click the link below to verify your email:",
				link,
				"This link will expire in 15 minutes.",
			)
		} else {
			email.PlainBody = "Verify your email address for Puchi\n\nYour verification code is: " + code + "\n\nThis code will expire in 15 minutes."
		}
	}
}

func actionLinkHTML(title, userName, intro, button, link string) string {
	return `<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f6f8fa; margin: 0; padding: 0;">
<table width="100%" cellpadding="0" cellspacing="0" style="background: #f6f8fa; padding: 40px 0;">
<tr><td align="center">
<table width="480" cellpadding="0" cellspacing="0" style="background: #ffffff; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.08);">
<tr><td style="padding: 40px 32px 24px;" align="center">
<img src="https://puchi.io.vn/logo.png" alt="Puchi" width="48" height="48" style="border-radius: 12px;" />
<h1 style="font-size: 20px; color: #1a1a2e; margin: 16px 0 4px;">Hi ` + userName + `!</h1>
<p style="color: #64748b; font-size: 14px; margin: 0;">` + intro + `</p>
<a href="` + link + `" style="display: inline-block; margin: 24px 0; padding: 12px 32px; background: #6366f1; color: #ffffff; text-decoration: none; border-radius: 8px; font-size: 15px; font-weight: 600;">` + button + `</a>
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

func emailVerifyCodeHTML(code, userName string) string {
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
