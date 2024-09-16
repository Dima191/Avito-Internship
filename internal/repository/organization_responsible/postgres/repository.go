package repository_organization_resp_postgres

import (
	"avito_intership/internal/repository"
	repository_organization_resp "avito_intership/internal/repository/organization_responsible"
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

func (r *rep) GetOrganizationIDByRepresentative(ctx context.Context, userID string) (organizationID string, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT organization_id FROM organization_responsible WHERE user_id = $1"

	if err = r.pool.QueryRow(ctx, stmt, userID).Scan(&organizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository_organization_resp.ErrUserHasNoOrganization
		}

		l.Error("Failed to get organization id by representative", "error", err.Error())
		return "", repository_organization_resp.ErrInternal
	}

	return organizationID, nil
}

func (r *rep) OrganizationRepresentativesAmount(ctx context.Context, organizationID string) (amount int, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT COUNT(*) FROM organization_responsible WHERE organization_id = $1"

	if err = r.pool.QueryRow(ctx, stmt, organizationID).Scan(&amount); err != nil {
		l.Error("Failed to count organization representatives", "error", err.Error())
		return 0, repository_organization_resp.ErrInternal
	}

	return amount, nil
}

func (r *rep) CloseConn() {
	r.pool.Close()
}

func New(ctx context.Context, connStr string, logger *slog.Logger) (repository_organization_resp.Repository, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		logger.Error("Failed to open connection to db", "error", err.Error())
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
