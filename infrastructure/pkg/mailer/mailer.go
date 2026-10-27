package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/wneessen/go-mail"
)

//go:embed templates/*.html
var embeddedTemplates embed.FS

// Client is a thin SMTP client with HTML template rendering.
type Client struct {
	templates map[string]*template.Template
	cfg       SMTPConfig
}

// SMTPConfig is the SMTP-specific subset of the global config.
type SMTPConfig struct {
	// use time.Duration alias
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	TLSMode     string
	Timeout     time.Duration
}

func NewClient(cfg *SMTPConfig) *Client {
	return &Client{
		templates: make(map[string]*template.Template),
		cfg:       *cfg,
	}
}

// EmailMessage is the payload for a single outbound email.
type EmailMessage struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	HTMLBody    string
	TextBody    string
	Attachments []string
}

// Send sends a plain or HTML email.
func (c *Client) Send(ctx context.Context, msg *EmailMessage) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("recipient is required")
	}

	m := mail.NewMsg()
	fromName := c.cfg.FromName
	fromEmail := c.cfg.Username
	if fromEmail2 := c.cfg.FromAddress; fromEmail2 != "" {
		fromEmail = fromEmail2
	}
	if err := m.FromFormat(fromName, fromEmail); err != nil {
		return err
	}

	if err := m.To(msg.To...); err != nil {
		return err
	}
	if len(msg.Cc) > 0 {
		if err := m.Cc(msg.Cc...); err != nil {
			return err
		}
	}
	if len(msg.Bcc) > 0 {
		if err := m.Bcc(msg.Bcc...); err != nil {
			return err
		}
	}

	m.Subject(msg.Subject)

	if msg.HTMLBody != "" {
		m.SetBodyString(mail.TypeTextHTML, msg.HTMLBody)
	} else {
		m.SetBodyString(mail.TypeTextPlain, msg.TextBody)
	}

	for _, a := range msg.Attachments {
		if _, err := os.Stat(a); err != nil {
			return fmt.Errorf("attachment not found: %s", a)
		}
		m.AttachFile(a)
	}

	timeout := c.cfg.Timeout
	if timeout <= 0 {
		timeout = mail.DefaultTimeout
	}

	smtp, err := mail.NewClient(
		c.cfg.Host,
		mail.WithPort(c.cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(c.cfg.Username),
		mail.WithPassword(c.cfg.Password),
		mail.WithTimeout(timeout),
		mail.WithTLSPolicy(c.tlsPolicy()),
	)
	if err != nil {
		return err
	}

	return smtp.DialAndSendWithContext(ctx, m)
}

// SendTemplate renders an embedded HTML template and sends it.
func (c *Client) SendTemplate(ctx context.Context, to []string, subject, templateName string, data any) error {
	if _, ok := c.templates[templateName]; !ok {
		raw, err := embeddedTemplates.ReadFile("templates/" + templateName)
		if err != nil {
			return err
		}
		tmpl, err := template.New(templateName).Parse(string(raw))
		if err != nil {
			return err
		}
		c.templates[templateName] = tmpl
	}

	var buf bytes.Buffer
	if err := c.templates[templateName].Execute(&buf, data); err != nil {
		return err
	}

	return c.Send(ctx, &EmailMessage{
		To:       to,
		Subject:  subject,
		HTMLBody: buf.String(),
	})
}

func (c *Client) tlsPolicy() mail.TLSPolicy {
	switch c.cfg.TLSMode {
	case "NONE", "DISABLE":
		return mail.TLSOpportunistic
	default:
		return mail.TLSMandatory
	}
}

// OtpData payload for otp_signup template.
type OtpData struct {
	Code    string `json:"-"`
	Expired int
}
