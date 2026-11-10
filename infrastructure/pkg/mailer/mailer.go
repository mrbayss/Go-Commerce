package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"os"
	"sync"
	"time"

	"github.com/wneessen/go-mail"
)

//go:embed templates/*.html
var embeddedTemplates embed.FS

// Client is a thin SMTP client with HTML template rendering.
type Client struct {
	templates  map[string]*template.Template
	templateMu sync.RWMutex
	cfg        SMTPConfig
	mailClient *mail.Client
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

func NewClient(cfg *SMTPConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("SMTP config is required")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = mail.DefaultTimeout
	}

	mailClient, err := mail.NewClient(
		cfg.Host,
		mail.WithPort(cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
		mail.WithTimeout(timeout),
		mail.WithTLSPolicy(tlsPolicy(cfg.TLSMode)),
	)
	if err != nil {
		return nil, fmt.Errorf("initialize SMTP client: %w", err)
	}

	return &Client{
		templates:  make(map[string]*template.Template),
		cfg:        *cfg,
		mailClient: mailClient,
	}, nil
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

	client := c.mailClient
	if client == nil {
		timeout := c.cfg.Timeout
		if timeout <= 0 {
			timeout = mail.DefaultTimeout
		}
		var err error
		client, err = mail.NewClient(
			c.cfg.Host,
			mail.WithPort(c.cfg.Port),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(c.cfg.Username),
			mail.WithPassword(c.cfg.Password),
			mail.WithTimeout(timeout),
			mail.WithTLSPolicy(tlsPolicy(c.cfg.TLSMode)),
		)
		if err != nil {
			return err
		}
	}

	return client.DialAndSendWithContext(ctx, m)
}

// SendTemplate renders an embedded HTML template and sends it.
func (c *Client) SendTemplate(ctx context.Context, to []string, subject, templateName string, data any) error {
	c.templateMu.RLock()
	tmpl, ok := c.templates[templateName]
	c.templateMu.RUnlock()

	if !ok {
		raw, err := embeddedTemplates.ReadFile("templates/" + templateName)
		if err != nil {
			return err
		}
		parsedTmpl, err := template.New(templateName).Parse(string(raw))
		if err != nil {
			return err
		}

		c.templateMu.Lock()
		c.templates[templateName] = parsedTmpl
		tmpl = parsedTmpl
		c.templateMu.Unlock()
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return c.Send(ctx, &EmailMessage{
		To:       to,
		Subject:  subject,
		HTMLBody: buf.String(),
	})
}

func tlsPolicy(mode string) mail.TLSPolicy {
	switch mode {
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
