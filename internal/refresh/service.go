package refresh

import (
	"context"

	"github.com/carloscfgos1980/taskSphere-api/internal/database"
	"github.com/jackc/pgx/v5"
)

// Service defines the interface for the users service
type Service interface {
	GetUserFromRefreshToken(ctx context.Context, refreshToken string) (database.User, error)
}

// svc defines the struct for the users service
type svc struct {
	repo *database.Queries
	db   *pgx.Conn
}

// NewService creates a new service for the users package
func NewService(repo *database.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

// GetUserFromRefreshToken retrieves the user associated with the given refresh token from the database
func (s *svc) GetUserFromRefreshToken(ctx context.Context, refreshToken string) (database.User, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.User{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)

	user, err := qtx.GetUserFromRefreshToken(ctx, refreshToken)
	if err != nil {
		return database.User{}, err
	}
	return user, nil
}
