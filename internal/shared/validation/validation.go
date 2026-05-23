package validation

import (
	"strings"
)

func Required(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return &FieldError{Field: field, Message: "is required"}
	}
	return nil
}

type FieldError struct {
	Field   string
	Message string
}

func (e *FieldError) Error() string {
	return e.Field + " " + e.Message
}
