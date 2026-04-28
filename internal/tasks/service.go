package tasks

import (
	"context"

	"github.com/carloscfgos1980/taskSphere-api/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service defines the interface for the users service
type Service interface {
	GetUserByID(ctx context.Context, id string) (database.User, error)
	CreateTask(ctx context.Context, task Task) (database.Task, error)
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

// GetUserByID retrieves a user by their ID from the database
func (s *svc) GetUserByID(ctx context.Context, id string) (database.User, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.User{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	userID := pgtype.UUID{}
	err = userID.Scan(id)
	if err != nil {
		return database.User{}, err
	}
	user, err := qtx.GetUserByID(ctx, userID)
	if err != nil {
		return database.User{}, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.User{}, err
	}
	return user, nil
}

// CreateTask creates a new task in the database associated with the given user ID
func (s *svc) CreateTask(ctx context.Context, task Task) (database.Task, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.Task{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)

	userID := pgtype.UUID{Bytes: task.UserID, Valid: true}

	endDate := pgtype.Timestamp{}
	err = endDate.Scan(task.EndDate)
	if err != nil {
		return database.Task{}, err
	}

	parentID := pgtype.UUID{Valid: false}
	if task.ParentID != uuid.Nil {
		parentID = pgtype.UUID{Bytes: task.ParentID, Valid: true}
	}

	taskEditors := make([]pgtype.UUID, len(task.TaskEditors))
	for i, id := range task.TaskEditors {
		taskEditors[i] = pgtype.UUID{Bytes: id, Valid: true}
	}

	createdTask, err := qtx.CreateTask(ctx, database.CreateTaskParams{
		UserID:      userID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		Tag:         task.Tag,
		State:       task.State,
		EndDate:     endDate,
		ParentID:    parentID,
		TaskEditors: taskEditors,
	})
	if err != nil {
		return database.Task{}, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.Task{}, err
	}
	return createdTask, nil
}
