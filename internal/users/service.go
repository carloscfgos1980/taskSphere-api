package users

import (
	"context"
	"time"

	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service defines the interface for the users service
type Service interface {
	CreateUser(ctx context.Context, user UserRequest) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	CreateRefreshToken(ctx context.Context, userID string, refreshToken string) error
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

// GetUserByEmail gets a user from the database by email
func (s *svc) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.User{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	// get the user by email
	user, err := qtx.GetUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.User{}, err
	}
	// return the user
	return user, nil

}

// CreateRefreshToken creates a new refresh token for a user in the database
func (s *svc) CreateRefreshToken(ctx context.Context, userID string, refreshToken string) error {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	// create the refresh token
	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	userUUID := pgtype.UUID{Bytes: parsedUUID, Valid: true}
	_, err = qtx.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		UserID:    userUUID,
		Token:     refreshToken,
		ExpiresAt: pgtype.Timestamp{Time: time.Now().UTC().Add(time.Hour * 24 * 60), Valid: true},
	})
	if err != nil {
		return err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
