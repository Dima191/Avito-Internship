package validator

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
)

const (
	AuthorTypeTag = "author_type"
)

var (
	ErrInvalidAuthorType = errors.New("invalid author type. possible accepted value: Organization, User")
)

type Validate struct {
	validate *validator.Validate
}

func (v *Validate) Validate(model any) (resErr error) {
	err := v.validate.Struct(model)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, validationErr := range validationErrors {
				switch validationErr.Tag() {
				case AuthorTypeTag:
					resErr = multierror.Append(resErr, ErrInvalidAuthorType)
				default:
					resErr = multierror.Append(resErr, err)
				}
			}
		}
		return resErr
	}

	return nil
}

func (v *Validate) RegisterTag(tag string, fn validator.Func) error {
	if err := v.validate.RegisterValidation(tag, fn); err != nil {
		return err
	}
	return nil
}

func New() *Validate {
	return &Validate{
		validate: validator.New(),
	}
}
