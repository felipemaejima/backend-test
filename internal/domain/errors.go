package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrPartNotFound = errors.New("peça não encontrada")

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", f.Field, f.Message))
	}
	return "validação falhou - " + strings.Join(parts, "; ")
}

func (e *ValidationError) add(field, message string) {
	e.Fields = append(e.Fields, FieldError{Field: field, Message: message})
}

func (e *ValidationError) orNil() error {
	if len(e.Fields) == 0 {
		return nil
	}
	return e
}
