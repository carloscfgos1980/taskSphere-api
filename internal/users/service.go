package users

import (
	"context"

	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/jackc/pgx/v5"
)

// Service defines the interface for the users service
type Service interface {
	CreateUser(ctx context.Context, user UserRequest) (database.User, error)
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

// CreateUser creates a new user in the database
func (s *svc) CreateUser(ctx context.Context, user UserRequest) (database.User, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.User{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)

	// create the user
	createdUser, err := qtx.CreateUser(ctx, database.CreateUserParams{
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return database.User{}, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.User{}, err
	}
	// return the created user
	return createdUser, nil
}
