package repository_decision_postgres

import (
	"avito_intership/internal/repository"
	repository_decision "avito_intership/internal/repository/decision"
	"avito_intership/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type rep struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func (r *rep) SubmitDecision(ctx context.Context, authorID string, tenderID string, bidID string, decision string) error {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "INSERT INTO decision(tender_author_id, tender_id, bid_id, decision) VALUES($1, $2, $3, $4)"

	_, err := r.pool.Exec(ctx, stmt, authorID, tenderID, bidID, decision)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == pgerrcode.UniqueViolation:
				return repository_decision.ErrUserAlreadyVoted
			case pgErr.Code == pgerrcode.ForeignKeyViolation:
				return repository_decision.ErrInvalidForeignKey
			}
		}

		l.Error("Failed to submit decision", "error", err.Error())
		return repository_decision.ErrInternal
	}

	return nil
}

func (r *rep) DecisionStats(ctx context.Context, bidID string) (applied int, rejected int, err error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := `SELECT COUNT(*) FILTER (WHERE decision = 'Approved') AS approvals, COUNT(*) FILTER (WHERE decision = 'Rejected') AS rejections FROM decision
	WHERE bid_id = $1`

	if err = r.pool.QueryRow(ctx, stmt, bidID).Scan(&applied, &rejected); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, repository_decision.ErrNoVotes
		}
		l.Error("Failed to get statistic by bid id", "error", err.Error())
		return 0, 0, repository_decision.ErrInternal
	}

	return applied, rejected, err
}

func (r *rep) CloseConn() {
	r.pool.Close()
}

func New(ctx context.Context, connStr string, logger *slog.Logger) (repository_decision.Repository, error) {
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
