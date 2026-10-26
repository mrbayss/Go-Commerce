package mail

import "github.com/wneessen/go-mail"

func resolveBody(msg *EmailMessage) (mail.ContentType, string) {
	if msg.HTMLBody != "" {
		return mail.TypeTextHTML, msg.HTMLBody
	}
	return mail.TypeTextPlain, msg.TextBody
}
