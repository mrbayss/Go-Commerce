package mail

import (
	"fmt"
	"os"

	"github.com/wneessen/go-mail"
)

func buildMessage(c *Client, msg *EmailMessage) (*mail.Msg, error) {
	if len(msg.To) == 0 {
		return nil, fmt.Errorf("recipient is required")
	}

	m := mail.NewMsg()

	if err := m.FromFormat(c.FromName, c.FromEmail); err != nil {
		return nil, err
	}

	if err := m.To(msg.To...); err != nil {
		return nil, err
	}

	if len(msg.Cc) > 0 {
		err := m.Cc(msg.Cc...)
		if err != nil {
			return nil, err
		}
	}
	if len(msg.Bcc) > 0 {
		err := m.Bcc(msg.Bcc...)
		if err != nil {
			return nil, err
		}
	}

	m.Subject(msg.Subject)

	ct, body := resolveBody(msg)
	m.SetBodyString(ct, body)

	for _, a := range msg.Attachments {
		if _, err := os.Stat(a); err != nil {
			return nil, fmt.Errorf("attachment not found: %s", a)
		}
		m.AttachFile(a)
	}

	return m, nil
}
