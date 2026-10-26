package mail

import (
	"context"
	"embed"
	"html/template"
	"time"

	"github.com/spf13/viper"
	"github.com/wneessen/go-mail"
)

type Client struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
	Timeout   time.Duration

	TLSPolicy   mail.TLSPolicy
	smtpFactory SMTPFactory
	templates   map[string]*template.Template
	templateFS  embed.FS
}

func NewClient(v *viper.Viper, fs embed.FS) (*Client, error) {

	timeout := mail.DefaultTimeout

	return &Client{
		Host:        v.GetString("smtp.host"),
		Port:        v.GetInt("smtp.port"),
		Username:    v.GetString("smtp.username"),
		Password:    v.GetString("smtp.password"),
		FromEmail:   v.GetString("smtp.username"),
		FromName:    v.GetString("smtp.from_name"),
		Timeout:     timeout,
		TLSPolicy:   mail.TLSMandatory,
		smtpFactory: defaultSMTPFactory,
		templates:   make(map[string]*template.Template),
		templateFS:  fs,
	}, nil
}

func (c *Client) Send(ctx context.Context, msg *EmailMessage) error {
	m, err := buildMessage(c, msg)
	if err != nil {
		return err
	}

	smtp, err := c.smtpFactory(c)
	if err != nil {
		return err
	}

	return smtp.DialAndSendWithContext(ctx, m)
}

func (c *Client) SendTemplate(ctx context.Context, to []string, subject, templateName string, data interface{}) error {
	if _, ok := c.templates[templateName]; !ok {
		tmpl, err := template.ParseFS(c.templateFS, templateName)
		if err != nil {
			return err
		}
		c.templates[templateName] = tmpl
	}

	body, err := RenderTemplate(c.templates[templateName], data)
	if err != nil {
		return err
	}

	return c.Send(ctx, &EmailMessage{
		To:       to,
		Subject:  subject,
		HTMLBody: body,
	})
}
