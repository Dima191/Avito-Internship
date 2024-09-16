package repository_employee_postgres

import (
	"avito_intership/internal/repository"
	repository_employee "avito_intership/internal/repository/employee"
	"avito_intership/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type rep struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func (r *rep) IDByUsername(ctx context.Context, username string) (userID string, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT id FROM employee WHERE username = $1"
	row := r.pool.QueryRow(ctx, stmt, username)
	if err = row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository_employee.ErrNonExistingEmployee
		}

		l.Error("Failed to get user_id by username", "error", err.Error())
		return "", repository_employee.ErrInternal
	}

	return userID, nil
}

func (r *rep) CloseConn() {
	r.pool.Close()
}

func New(ctx context.Context, connectionStr string, logger *slog.Logger) (repository_employee.Repository, error) {
	pool, err := pgxpool.New(ctx, connectionStr)
	if err != nil {
		logger.Error("Failed to create connection to db", "error", err.Error())
		return nil, repository.ErrOpenConn
	}

	if err = pool.Ping(ctx); err != nil {
		logger.Error("Failed to ping db", "error", err.Error())
		return nil, repository.ErrPingDB
	}

	r := &rep{
		pool:   pool,
		logger: logger,
	}

	return r, nil
}
