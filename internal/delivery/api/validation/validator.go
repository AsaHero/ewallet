package validation

import (
	"github.com/go-playground/validator/v10"

)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	validator := validator.New()

	return &Validator{
		validator: validator,
	}
}

func (v Validator) Validate(data any) error {
	return v.validator.Struct(data)
}
