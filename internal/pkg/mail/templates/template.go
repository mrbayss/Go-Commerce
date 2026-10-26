package templates

import (
	"embed"
)

//go:embed *.html
var FS embed.FS

type OtpData struct {
	Code    string
	Expired int
}
