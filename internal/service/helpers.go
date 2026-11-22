package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// serviceLogger builds a request-scoped log entry tagged with the request id
// and the calling function.
func serviceLogger(ctx context.Context, log *logrus.Logger, functionName string) *logrus.Entry {
	reqID, _ := ctx.Value("request_id").(string)
	return log.WithFields(logrus.Fields{
		"request_id": reqID,
		"function":   functionName,
	})
}

// validateStruct validates request and logs failures with the given entry.
func validateStruct(log *logrus.Entry, validation *validator.Validate, request any) error {
	if err := validation.Struct(request); err != nil {
		log.WithError(err).Error("validation failed")
		return err
	}
	return nil
}
