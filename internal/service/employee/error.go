package service_employee

import "errors"

var (
	ErrInternal            = errors.New("internal error")
	ErrNonExistingEmployee = errors.New("non existing employee")
)
