package storage

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const PublicPrefix = "/uploads"

func Save(fh *multipart.FileHeader, dir, filename string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func Remove(path string) error {
	return os.Remove(path)
}

func DiskPath(uploadDir, publicURL string) string {
	return filepath.Join(uploadDir, strings.TrimPrefix(publicURL, PublicPrefix))
}
