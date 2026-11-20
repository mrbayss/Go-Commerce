package storage

const (
	MIMEJPEG = "image/jpeg"
	MIMEPNG  = "image/png"
	MIMEWebP = "image/webp"
	MIMEGIF  = "image/gif"
)

var mimeExtensions = map[string]string{
	MIMEJPEG: ".jpg",
	MIMEPNG:  ".png",
	MIMEWebP: ".webp",
	MIMEGIF:  ".gif",
}
