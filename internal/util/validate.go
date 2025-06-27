package util

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)
// TODO: add checking errors
func NewCustomValidator() *validator.Validate {
	validate := validator.New()
	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if len(value) < 5 || len(value) > 30 {
			return false
		}
		return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value)
	})

	_ = validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if len(value) < 8 {
			return false
		}
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(value)
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(value)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(value)
		return hasNumber && hasUpper && hasLower
	})

	return validate
}
