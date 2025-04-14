package validator

import (
	"strings"
	"unicode/utf8"
)

// Validator - define a validator struct which contains a map of validation errors
type Validator struct {
	FieldErros map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.FieldErros) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErros == nil {
		v.FieldErros = make(map[string]string)
	}

	if _, exists := v.FieldErros[key]; !exists {
		v.FieldErros[key] = message
	}
}
func (v *Validator) ChecckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank returns true if a value contains no more than n characters.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars returns true if a value contains no more than n characters.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// PermittedInt returns true if a value is in the list of permitted integers.
func PermittedInt(value int, permittedValues ...int) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}
