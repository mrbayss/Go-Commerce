package mail

import (
	"context"

	"github.com/wneessen/go-mail"
)

type SMTPClient interface {
	DialAndSendWithContext(ctx context.Context, messages ...*mail.Msg) error
}
type SMTPFactory func(c *Client) (SMTPClient, error)

func defaultSMTPFactory(c *Client) (SMTPClient, error) {
	return mail.NewClient(
		c.Host,
		mail.WithPort(c.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(c.Username),
		mail.WithPassword(c.Password),
		mail.WithTimeout(c.Timeout),
		mail.WithTLSPolicy(c.TLSPolicy),
	)
}
