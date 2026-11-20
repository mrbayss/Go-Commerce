package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
)

var ErrUnsupportedType = errors.New("unsupported file type")

func SaveUpload(fh *multipart.FileHeader, uploadDir, subdir string, mimes ...string) (string, error) {
	mime, err := sniffMIME(fh)
	if err != nil {
		return "", err
	}

	allowed := false
	for _, m := range mimes {
		if _, ok := mimeExtensions[m]; !ok {
			panic(fmt.Sprintf("storage: unknown MIME type %q", m))
		}
		if m == mime {
			allowed = true
		}
	}
	if !allowed {
		return "", ErrUnsupportedType
	}

	filename := uuid.NewString() + mimeExtensions[mime]
	if err := Save(fh, filepath.Join(uploadDir, subdir), filename); err != nil {
		return "", err
	}
	return PublicPrefix + "/" + subdir + "/" + filename, nil
}

func sniffMIME(fh *multipart.FileHeader) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := io.ReadFull(src, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}
