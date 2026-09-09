package router

import "gopkg.in/go-playground/validator.v9"

func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// gin has no per-engine validator hook to hang this off, the way echo's
// Echo.Validator did, so request binding calls Validate directly.
var defaultValidator = NewValidator()

func Validate(i interface{}) error {
	return defaultValidator.Validate(i)
}
