package validator

import (
	"fmt"
	"math"

	"github.com/go-playground/validator/v10"
	"github.com/muriloabranches/fintrack-api/internal/domain"
)

// validate is the package-level validator instance.
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ToCents converts a decimal amount (max 2 decimal places) to cents (int64).
// Returns domain.ErrInvalidAmount if the value has more than 2 decimal places.
func ToCents(amount float64) (int64, error) {
	rounded := math.Round(amount * 100)
	if math.Abs(amount*100-rounded) > 0.001 {
		return 0, domain.ErrInvalidAmount
	}
	return int64(rounded), nil
}

// FromCents converts a cents value (int64) to a decimal float64.
func FromCents(cents int64) float64 {
	return float64(cents) / 100.0
}

// Validate validates a struct using go-playground/validator tags.
// Returns nil if the struct is valid, or a *domain.ValidationError with field details.
func Validate(s interface{}) *domain.ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return &domain.ValidationError{
			Fields: []domain.FieldError{
				{Field: "unknown", Message: "validation failed"},
			},
		}
	}

	fields := make([]domain.FieldError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		fields = append(fields, domain.FieldError{
			Field:   fieldName(fe),
			Message: friendlyMessage(fe),
		})
	}

	return &domain.ValidationError{Fields: fields}
}

// fieldName returns a lowercase field name from the validator FieldError.
func fieldName(fe validator.FieldError) string {
	return fe.Field()
}

// friendlyMessage maps validator tags to user-friendly messages.
func friendlyMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	default:
		return fmt.Sprintf("failed on '%s' validation", fe.Tag())
	}
}
