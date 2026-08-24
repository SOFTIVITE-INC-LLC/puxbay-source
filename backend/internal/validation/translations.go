package validation

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	uni      *ut.UniversalTranslator
	Validate *validator.Validate
	Trans    ut.Translator
)

// InitTranslations sets up custom validation messages.
func InitTranslations() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		enLocale := en.New()
		uni = ut.New(enLocale, enLocale)
		Trans, _ = uni.GetTranslator("en")

		// Register default english translations
		en_translations.RegisterDefaultTranslations(v, Trans)

		// Register custom JSON tag extraction for field names
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		Validate = v
	}
}

// FormatValidationErrors translates validation errors into a map of field -> message
func FormatValidationErrors(err error) map[string]string {
	errs := make(map[string]string)

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			errs[e.Field()] = e.Translate(Trans)
		}
	} else {
		errs["error"] = err.Error()
	}

	return errs
}
