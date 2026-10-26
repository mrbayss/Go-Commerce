package mail

type EmailMessage struct {
	To          []string `validate:"required" json:"to"`
	Cc          []string `validate:"" json:"cc"`
	Bcc         []string `validate:"" json:"bcc"`
	Subject     string   `validate:"required" json:"subject"`
	HTMLBody    string   `validate:"required" json:"htmlBody"`
	TextBody    string   `validate:"" json:"textBody"`
	Attachments []string `validate:"" json:"attachments"`
}
