package service_employee

import "context"

type Service interface {
	IDByUsername(ctx context.Context, username string) (userID string, err error)
}
