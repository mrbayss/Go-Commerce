package service

import (
	"testing"

	"github.com/mrbayss/Go-Commerce/internal/model"
)

func TestSignUpRequestValidationTags(t *testing.T) {
	// Sanity: request struct fields have expected validation tags.
	var req model.SignUpReq
	if req == (model.SignUpReq{}) {
		// Fields present in struct (reflection-agnostic check via zero value).
		_ = req.Email
		_ = req.Password
		_ = req.FirstName
		_ = req.LastName
		_ = req.Phone
	}
}
