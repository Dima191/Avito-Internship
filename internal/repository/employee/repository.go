package repository_employee

import "context"

type Repository interface {
	IDByUsername(ctx context.Context, username string) (userID string, err error)
	CloseConn()
}
