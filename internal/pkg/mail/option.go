package mail

type Option func(*EmailMessage)

func WithCC(cc ...string) Option {
	return func(m *EmailMessage) {
		m.Cc = append(m.Cc, cc...)
	}
}

func WithBCC(bcc ...string) Option {
	return func(m *EmailMessage) {
		m.Bcc = append(m.Bcc, bcc...)
	}
}

func WithAttachment(paths ...string) Option {
	return func(m *EmailMessage) {
		m.Attachments = append(m.Attachments, paths...)
	}
}

func WithTextBody(text string) Option {
	return func(m *EmailMessage) {
		m.TextBody = text
	}
}
