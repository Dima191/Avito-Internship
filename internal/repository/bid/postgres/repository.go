package repository_bid_postgres

import (
	"avito_intership/internal/model"
	"avito_intership/internal/repository"
	repository_bid "avito_intership/internal/repository/bid"
	repository_bid_converter "avito_intership/internal/repository/bid/converter"
	repository_bid_model "avito_intership/internal/repository/bid/model"
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

var (
	authorIDRaiseExceptionMsg = "author_id"
	tenderIDConstraint        = "bid_tender_id_fkey"
)

func (r *rep) Create(ctx context.Context, bid model.Bid) (model.Bid, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	repositoryBid := repository_bid_model.Bid{}

	stmt := `INSERT INTO bid (name, description, tender_id, author_type, author_id) VALUES ($1,$2,$3,$4,$5) 
RETURNING id, name, status, author_type, author_id, version, created_at`
	row := r.pool.QueryRow(ctx, stmt,
		*bid.Name,
		*bid.Description,
		*bid.TenderID,
		*bid.AuthorType,
		*bid.AuthorID)

	if err := row.Scan(&repositoryBid.ID,
		&repositoryBid.Name,
		&repositoryBid.Status,
		&repositoryBid.AuthorType,
		&repositoryBid.AuthorID,
		&repositoryBid.Version,
		&repositoryBid.CreatedAt); err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == pgerrcode.InvalidTextRepresentation:
				return model.Bid{}, repository_bid.ErrInvalidReq
			case pgErr.Code == pgerrcode.ForeignKeyViolation:
				if pgErr.ConstraintName == tenderIDConstraint {
					return model.Bid{}, repository_bid.ErrInvalidTenderID
				}
			case pgErr.Code == pgerrcode.RaiseException:
				if strings.Contains(pgErr.Message, authorIDRaiseExceptionMsg) {
					return model.Bid{}, repository_bid.ErrInvalidAuthorID
				}
			}
		}

		l.Error("Failed to create bid", "error", err.Error())
		return model.Bid{}, repository_bid.ErrInternal
	}

	return repository_bid_converter.ToBidFromRepository(repositoryBid), nil
}

func (r *rep) BidsByAuthorID(ctx context.Context, userID, organizationID string, limit int, offset int) ([]model.Bid, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := `SELECT id, name, status, author_type, author_id, version, created_at FROM bid
                                                                  WHERE author_id IN ($1, $2) ORDER BY name LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, stmt, userID, organizationID, limit, offset)
	if err != nil {
		l.Error("Failed to get list of user bid", "error", err.Error())
		return nil, repository_bid.ErrInternal
	}
	defer rows.Close()

	bids := make([]model.Bid, 0, limit)

	for rows.Next() {
		bid := repository_bid_model.Bid{}
		if err = rows.Scan(&bid.ID,
			&bid.Name,
			&bid.Status,
			&bid.AuthorType,
			&bid.AuthorID,
			&bid.Version,
			&bid.CreatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository_bid.ErrNoBids
			}

			l.Error("Failed to get list of user bid", "error", err.Error())
			return nil, repository_bid.ErrInternal
		}

		bids = append(bids, repository_bid_converter.ToBidFromRepository(bid))
	}

	if err = rows.Err(); err != nil {
		l.Error("Failed to get list of user bid", "error", err.Error())
		return nil, repository_bid.ErrInternal
	}

	return bids, nil
}

func (r *rep) BidsByTenderID(ctx context.Context, tenderID string, limit int, offset int) (bids []model.Bid, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := `SELECT id, name, status, author_type, author_id, version, created_at FROM bid
                                                                  WHERE tender_id = $1 LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, stmt, tenderID, limit, offset)
	if err != nil {
		l.Error("Failed to get bid by tender_id", "error", err.Error())
		return nil, repository_bid.ErrInternal
	}
	defer rows.Close()

	bids = make([]model.Bid, 0, limit)

	for rows.Next() {
		bid := repository_bid_model.Bid{}
		if err = rows.Scan(&bid.ID,
			&bid.Name,
			&bid.Status,
			&bid.AuthorType,
			&bid.AuthorID,
			&bid.Version,
			&bid.CreatedAt); err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository_bid.ErrNoBids
			}

			l.Error("Failed to get bid by tender_id", "error", err.Error())
			return nil, repository_bid.ErrInternal
		}

		bids = append(bids, repository_bid_converter.ToBidFromRepository(bid))
	}

	if err = rows.Err(); err != nil {
		l.Error("Failed to get bid by tender_id", "error", err.Error())
		return nil, repository_bid.ErrInternal
	}

	return bids, nil
}

func (r *rep) GetStatus(ctx context.Context, bidID string) (status string, tenderID string, authorID string, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)
	stmt := "SELECT status, tender_id, author_id  FROM bid WHERE id = $1"
	if err = r.pool.QueryRow(ctx, stmt, bidID).Scan(&status, &tenderID, &authorID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", repository_bid.ErrNoBids
		}
		l.Error("Failed to get bid status", "error", err.Error())
		return "", "", "", repository_bid.ErrInternal
	}

	return status, tenderID, authorID, nil
}

func (r *rep) ChangeStatus(ctx context.Context, bidID string, status string) (bid model.Bid, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	repositoryBid := repository_bid_model.Bid{}

	stmt := "UPDATE bid SET status = $1 WHERE id = $2 RETURNING id, name, status, author_type, author_id, version, created_at"
	if err = r.pool.QueryRow(ctx, stmt, status, bidID).Scan(&repositoryBid.ID,
		&repositoryBid.Name,
		&repositoryBid.Status,
		&repositoryBid.AuthorType,
		&repositoryBid.AuthorID,
		&repositoryBid.Version,
		&repositoryBid.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Bid{}, repository_bid.ErrNoBids
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == pgerrcode.InvalidTextRepresentation:
				return model.Bid{}, repository_bid.ErrInvalidBidStatus
			}
		}

		l.Error("Failed to change bid status", "error", err.Error())
		return model.Bid{}, repository_bid.ErrInternal
	}

	return repository_bid_converter.ToBidFromRepository(repositoryBid), nil
}

func (r *rep) Edit(ctx context.Context, bidID string, bid model.Bid) (updatedBid model.Bid, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	sqlPatch := sql_patch.SQLPatches(bid)

	if len(sqlPatch.Args) == 0 {
		return model.Bid{}, repository_bid.ErrNoSuggestionToUpdate
	}

	stmt := fmt.Sprintf(`UPDATE bid SET %s WHERE id = $%d
                   RETURNING id, name, status, author_type, author_id, version, created_at`, strings.Join(sqlPatch.Fields, ", "), len(sqlPatch.Args)+1)

	repositoryBid := repository_bid_model.Bid{}

	row := r.pool.QueryRow(ctx, stmt, append(sqlPatch.Args, bidID)...)
	if err = row.Scan(&repositoryBid.ID,
		&repositoryBid.Name,
		&repositoryBid.Status,
		&repositoryBid.AuthorType,
		&repositoryBid.AuthorID,
		&repositoryBid.Version,
		&repositoryBid.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Bid{}, repository_bid.ErrNoBids
		}

		l.Error("Failed to edit bid", "error", err.Error())
		return model.Bid{}, repository_bid.ErrInternal
	}

	return repository_bid_converter.ToBidFromRepository(repositoryBid), nil
}

func (r *rep) BidTenderID(ctx context.Context, bidID string) (tenderID string, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT tender_id FROM bid WHERE id = $1"

	if err = r.pool.QueryRow(ctx, stmt, bidID).Scan(&tenderID); err != nil {
		l.Error("Failed to get organization id", "error", err.Error())
		return "", repository_bid.ErrInternal
	}

	return tenderID, nil
}

func (r *rep) BidAuthorID(ctx context.Context, bidID string) (authorID string, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT author_id FROM bid WHERE id = $1"

	if err = r.pool.QueryRow(ctx, stmt, bidID).Scan(&authorID); err != nil {
		l.Error("Failed to get organization id", "error", err.Error())
		return "", repository_bid.ErrInternal
	}

	return authorID, nil
}

func (r *rep) BidByID(ctx context.Context, bidID string) (bid model.Bid, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "SELECT id, name, status, author_type, author_id, version, created_at FROM bid WHERE id = $1"

	if err = r.pool.QueryRow(ctx, stmt, bidID).Scan(&bid.ID,
		&bid.Name,
		&bid.Status,
		&bid.AuthorType,
		&bid.AuthorID,
		&bid.Version,
		&bid.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Bid{}, repository_bid.ErrNoBids
		}

		l.Error("Failed to get bid by id", "error", err.Error())
		return model.Bid{}, repository_bid.ErrInternal
	}
	return bid, nil
}

func (r *rep) RollbackVersion(ctx context.Context, bidID string, version int) (model.Bid, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := `UPDATE bid
	SET name = bh.name,
		description = bh.description,
		status = bh.status,
		tender_id = bh.tender_id,
		author_type = bh.author_type,
		author_id = bh.author_id,
		version = bh.version,
		created_at = bh.created_at
	FROM bid_history bh
	WHERE bid.id = bh.id AND bh.id = $1 AND bh.version = $2
	RETURNING bh.id, bh.name, bh.status, bh.author_type, bh.author_id, bh.version, bh.created_at`

	repositoryBid := repository_bid_model.Bid{}

	row := r.pool.QueryRow(ctx, stmt, bidID, version)
	if err := row.Scan(&repositoryBid.ID,
		&repositoryBid.Name,
		&repositoryBid.Status,
		&repositoryBid.AuthorType,
		&repositoryBid.AuthorID,
		&repositoryBid.Version,
		&repositoryBid.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Bid{}, repository_bid.ErrNoBids
		}
		l.Error("Failed to rollback version", "error", err.Error())
		return model.Bid{}, repository_bid.ErrInternal
	}
	return repository_bid_converter.ToBidFromRepository(repositoryBid), nil
}

func (r *rep) CloseConn() {
	r.pool.Close()
}

func New(ctx context.Context, connStr string, logger *slog.Logger) (repository_bid.Repository, error) {
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
