package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatErrors(err error) (bool, string) {
	errors := make([]string, 0)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, e.Error())
		}
	}

	if len(errors) == 0 {
		return true, ""
	}

	return false, strings.Join(errors, "; ")
}
