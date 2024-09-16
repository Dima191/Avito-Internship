package repository_feedback_postgres

import (
	"avito_intership/internal/model"
	"avito_intership/internal/repository"
	repository_feedback "avito_intership/internal/repository/feedback"
	repository_feedback_converter "avito_intership/internal/repository/feedback/converter"
	repository_feedback_model "avito_intership/internal/repository/feedback/model"
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

func (r *rep) Feedback(ctx context.Context, userID string, feedback string) error {
	l := logger.EndToEndLogging(ctx, r.logger)

	stmt := "INSERT INTO review(description, author_username) VALUES ($1, $2)"
	if _, err := r.pool.Exec(ctx, stmt, feedback, userID); err != nil {
		l.Error("Failed to create feedback", "error", err.Error())
		return repository_feedback.ErrInternal
	}
	return nil
}

func (r *rep) GetFeedbacks(ctx context.Context, authorUsername string, limit int, offset int) ([]model.Feedback, error) {
	l := logger.EndToEndLogging(ctx, r.logger)

	feedbacks := make([]model.Feedback, 0, limit)

	stmt := "SELECT id, description, created_at FROM review WHERE author_username = $1 LIMIT $2 OFFSET $3"

	rows, err := r.pool.Query(ctx, stmt, authorUsername, limit, offset)
	if err != nil {
		l.Error("Failed to get review", "error", err.Error())
		return nil, repository_feedback.ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		feedback := repository_feedback_model.Feedback{}
		if err = rows.Scan(&feedback.ID, &feedback.Description, &feedback.CreatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository_feedback.ErrNoReviews
			}

			l.Error("Failed to get review", "error", err.Error())
			return nil, repository_feedback.ErrInternal
		}

		feedbacks = append(feedbacks, repository_feedback_converter.ToFeedbackFromRepository(feedback))
	}

	if err = rows.Err(); err != nil {
		l.Error("Failed to get review", "error", err.Error())
		return nil, repository_feedback.ErrInternal
	}

	return feedbacks, err
}

func (r *rep) CloseConn() {
	r.pool.Close()
}

func New(ctx context.Context, connStr string, logger *slog.Logger) (repository_feedback.Repository, error) {
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
