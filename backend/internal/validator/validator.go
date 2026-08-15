package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Init builds the package-level validator instance — call once at startup,
// before any ValidateStruct call. Custom validation tags (beyond the
// built-in ones like "required") get registered here too, once there are
// any domain-specific rules that need them.
func Init() {
	validate = validator.New()
}

// ValidateStruct checks s against its own `validate:"..."` struct tags.
func ValidateStruct(s any) error {
	return validate.Struct(s)
}
