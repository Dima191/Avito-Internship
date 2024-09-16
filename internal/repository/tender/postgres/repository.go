package repository_tenders_postgres

import (
	"avito_intership/internal/model"
	"avito_intership/internal/repository"
	repository_tenders "avito_intership/internal/repository/tender"
	repository_tender_converter "avito_intership/internal/repository/tender/converter"
	repository_tender_model "avito_intership/internal/repository/tender/model"
	"avito_intership/pkg/logger"
	"avito_intership/pkg/sql_patch"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strings"
)

type rep struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func (r *rep) TenderOrganizationID(ctx context.Context, tenderID string) (organizationID string, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT organization_id FROM tender WHERE id = $1"

	if err = r.pool.QueryRow(ctx, stmt, tenderID).Scan(&organizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository_tenders.ErrNoTenders
		}
		l.Error("Failed to get tender organization id", "error", err.Error())
		return "", repository_tenders.ErrInternal
	}

	return organizationID, nil
}

func (r *rep) TenderList(ctx context.Context, serviceTypes []string, limit int, offset int) ([]model.Tender, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	args := make([]interface{}, 0)
	var condition string
	switch {
	case serviceTypes != nil:
		condition = "WHERE service_type IN ("
		for index, serviceType := range serviceTypes {
			condition = fmt.Sprintf("%s $%d", condition, index+1)
			if index != len(serviceTypes)-1 {
				condition += ","
			}
			args = append(args, serviceType)
		}
		condition = fmt.Sprintf("%s) LIMIT $%d OFFSET $%d", condition, len(serviceTypes)+1, len(serviceTypes)+2)
		args = append(args, limit, offset)
	default:
		condition = "LIMIT $1 OFFSET $2"
		args = append(args, limit, offset)
	}

	stmt := fmt.Sprintf("SELECT id, name, description, status, service_type, version, created_at FROM tender %s", condition)

	rows, err := r.pool.Query(ctx, stmt, args...)
	if err != nil {
		l.Error("Failed to get tender list", "error", err.Error())
		return nil, repository_tenders.ErrInternal
	}
	defer rows.Close()

	tenders := make([]model.Tender, 0, limit)

	for rows.Next() {
		tender := repository_tender_model.Tender{}
		if err = rows.Scan(&tender.ID,
			&tender.Name,
			&tender.Description,
			&tender.Status,
			&tender.ServiceType,
			&tender.Version,
			&tender.CreatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository_tenders.ErrNoTenders
			}
			l.Error("Failed to get tender list", "error", err.Error())
			return nil, repository_tenders.ErrInternal
		}
		tenders = append(tenders, repository_tender_converter.ToTenderFromRepository(tender))
	}

	if err = rows.Err(); err != nil {
		l.Error("Failed to get tender list", "error", err.Error())
		return nil, repository_tenders.ErrInternal
	}

	return tenders, nil
}

func (r *rep) Create(ctx context.Context, tender model.Tender) (model.Tender, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := `INSERT INTO tender (name, description, service_type, organization_id, creator_username) VALUES ($1,$2,$3,$4,$5)
RETURNING id, name, description, status, service_type, version, created_at`

	row := r.pool.QueryRow(ctx, stmt, tender.Name, tender.Description, tender.ServiceType, tender.OrganizationID, tender.CreatorUsername)

	repoTender := repository_tender_model.Tender{}
	if err := row.Scan(&repoTender.ID,
		&repoTender.Name,
		&repoTender.Description,
		&repoTender.Status,
		&repoTender.ServiceType,
		&repoTender.Version,
		&repoTender.CreatedAt); err != nil {
		l.Error("Failed to create tender", "error", err.Error())
		return model.Tender{}, repository_tenders.ErrInternal
	}

	return repository_tender_converter.ToTenderFromRepository(repoTender), nil
}

func (r *rep) TendersByUser(ctx context.Context, username string, limit int, offset int) ([]model.Tender, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT id, name, description, status, service_type, version, created_at FROM tender WHERE creator_username = $1 LIMIT $2 OFFSET $3"

	rows, err := r.pool.Query(ctx, stmt, username, limit, offset)
	if err != nil {
		l.Error("Failed to get tender list by user", "error", err.Error())
		return nil, repository_tenders.ErrInternal
	}
	defer rows.Close()

	tenders := make([]model.Tender, 0)

	for rows.Next() {
		tender := repository_tender_model.Tender{}
		if err = rows.Scan(&tender.ID,
			&tender.Name,
			&tender.Description,
			&tender.Status,
			&tender.ServiceType,
			&tender.Version,
			&tender.CreatedAt); err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository_tenders.ErrNoTenders
			}

			l.Error("Failed to get tender list by user", "error", err.Error())
			return nil, repository_tenders.ErrInternal
		}
		tenders = append(tenders, repository_tender_converter.ToTenderFromRepository(tender))
	}

	if err = rows.Err(); err != nil {
		l.Error("Failed to get tender list by user", "error", err.Error())
		return nil, repository_tenders.ErrInternal
	}

	return tenders, nil
}

func (r *rep) TenderStatus(ctx context.Context, tenderID string) (tenderOrganizationID string, status string, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT organization_id, status FROM tender WHERE id = $1"

	if err = r.pool.QueryRow(ctx, stmt, tenderID).Scan(&tenderOrganizationID, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", repository_tenders.ErrNoTenders
		}

		l.Error("Failed to get tender status by tender id", "error", err.Error())
		return "", "", repository_tenders.ErrInternal
	}

	return tenderOrganizationID, status, nil
}

func (r *rep) ChangeTenderStatusWithUserCheck(ctx context.Context, tenderID string, username string, status string) (model.Tender, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := `UPDATE tender SET status = $1 WHERE id = $2 AND creator_username = $3
RETURNING id, name, description, status, service_type, version, created_at`

	tender := repository_tender_model.Tender{}
	if err := r.pool.QueryRow(ctx, stmt, status, tenderID, username).Scan(&tender.ID,
		&tender.Name,
		&tender.Description,
		&tender.Status,
		&tender.ServiceType,
		&tender.Version,
		&tender.CreatedAt); err != nil {

		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr):
			switch {
			case pgErr.Code == pgerrcode.InvalidTextRepresentation:
				return model.Tender{}, repository_tenders.ErrInvalidStatus
			}
		case errors.Is(err, sql.ErrNoRows):
			return model.Tender{}, repository_tenders.ErrNoTenders
		default:
			l.Error("Failed to change tender status", "error", err.Error())
			return model.Tender{}, repository_tenders.ErrInternal
		}
	}

	return repository_tender_converter.ToTenderFromRepository(tender), nil
}

func (r *rep) ChangeTenderStatusForce(ctx context.Context, tenderID string, status string) error {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := `UPDATE tender SET status = $1 WHERE id = $2
RETURNING id, name, description, status, service_type, version, created_at`

	if _, err := r.pool.Exec(ctx, stmt, status, tenderID); err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr):
			switch {
			case pgErr.Code == pgerrcode.InvalidTextRepresentation:
				return repository_tenders.ErrInvalidStatus
			}
		case errors.Is(err, sql.ErrNoRows):
			return repository_tenders.ErrNoTenders
		default:
			l.Error("Failed to change tender status", "error", err.Error())
			return repository_tenders.ErrInternal
		}
	}

	return nil
}

func (r *rep) Edit(ctx context.Context, tenderID string, tender model.Tender) (model.Tender, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	sqlPatch := sql_patch.SQLPatches(tender)

	if len(sqlPatch.Args) == 0 {
		return model.Tender{}, repository_tenders.ErrNoSuggestionToUpdate
	}

	stmt := fmt.Sprintf(`UPDATE tender SET %s WHERE id = $1
	                   RETURNING id, name, description, status, service_type, version, created_at`, strings.Join(sqlPatch.Fields, ", "))

	repositoryTender := repository_tender_model.Tender{}

	row := r.pool.QueryRow(ctx, stmt, append(sqlPatch.Args, tenderID))
	if err := row.Scan(&repositoryTender.ID,
		&repositoryTender.Name,
		&repositoryTender.Description,
		&repositoryTender.Status,
		&repositoryTender.ServiceType,
		&repositoryTender.Version,
		&repositoryTender.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Tender{}, repository_tenders.ErrNoTenders
		}

		l.Error("Failed to edit tender", "error", err.Error())
		return model.Tender{}, repository_tenders.ErrInternal
	}

	return repository_tender_converter.ToTenderFromRepository(repositoryTender), nil
}

func (r *rep) RollbackVersion(ctx context.Context, tenderID string, version int) (model.Tender, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := `UPDATE tender
		SET name = th.name,
			description = th.description,
			service_type = th.service_type,
			status = th.status,
			organization_id = th.organization_id,
			creator_username = th.creator_username,
			version = th.version,
			created_at = th.created_at
		FROM tender_history th
		WHERE tender.id = th.id AND th.id = $1 AND th.version = $2
		RETURNING th.id, th.name, th.description, th.status, th.service_type, th.version, th.created_at`

	repositoryTender := repository_tender_model.Tender{}
	if err := r.pool.QueryRow(ctx, stmt, tenderID, version).Scan(&repositoryTender.ID,
		&repositoryTender.Name,
		&repositoryTender.Description,
		&repositoryTender.Status,
		&repositoryTender.ServiceType,
		&repositoryTender.Version,
		&repositoryTender.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Tender{}, repository_tenders.ErrNoTenders
		}

		l.Error("Failed to rollback version", "error", err.Error())
		return model.Tender{}, repository_tenders.ErrInternal
	}
	return repository_tender_converter.ToTenderFromRepository(repositoryTender), nil
}

func (r *rep) ConfirmTenderCreator(ctx context.Context, tenderID string, userOrganizationID string) (exists bool, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT EXISTS(SELECT 1 FROM tender WHERE id = $1 and organization_id = $2)"

	if err = r.pool.QueryRow(ctx, stmt, tenderID, userOrganizationID).Scan(&exists); err != nil {
		l.Error("Failed to confirm tender creator", "error", err.Error())
		return exists, repository_tenders.ErrInternal
	}

	return exists, nil
}

func (r *rep) CloseConn() {
	r.pool.Close()
}

func New(ctx context.Context, connStr string, logger *slog.Logger) (repository_tenders.Repository, error) {
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
