package response

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type WebResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Errors any    `json:"errors,omitempty"`
}

func FormatValidationError(err error) map[string]string {
	formattedErrors := make(map[string]string)

	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				formattedErrors[field] = field + " wajib diisi"
			case "min":
				formattedErrors[field] = field + " minimal " + e.Param() + " karakter/item"
			default:
				formattedErrors[field] = "Format " + field + " tidak valid"
			}
		}
		return formattedErrors
	}

	return map[string]string{"general": err.Error()}
}
