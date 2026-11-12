package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	// Cara terefisien: daftarkan parser JSON tag HANYA sekali saat inisialisasi
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

func (v *Validator) Validate(i interface{}) map[string]string {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	// Safety check dari kodemu (menghindari panic)
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{
			"error": err.Error(),
		}
	}

	errors := make(map[string]string)
	for _, e := range validationErrors {
		// e.Field() sekarang OTOMATIS berisi nama JSON tag berkat RegisterTagNameFunc
		errors[e.Field()] = humanizeError(e)
	}
	return errors
}

func humanizeError(e validator.FieldError) string {
	param := e.Param()
	
	switch e.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if e.Kind() == reflect.String {
			return fmt.Sprintf("minimal %s karakter", param)
		}
		return fmt.Sprintf("minimal %s", param)
	case "max":
		if e.Kind() == reflect.String {
			return fmt.Sprintf("maksimal %s karakter", param)
		}
		return fmt.Sprintf("maksimal %s", param)
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", param)
	case "uuid4", "uuid":
		return "format UUID tidak valid"
	case "len":
		return fmt.Sprintf("harus tepat %s karakter", param)
	case "numeric":
		return "harus berupa angka"
	default:
		return "tidak valid"
	}
}