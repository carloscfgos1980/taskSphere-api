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
	GetTaskByID(ctx context.Context, id string) (database.Task, error)
	GetTasksByUserID(ctx context.Context, userID string) ([]database.Task, error)
	GetCollaborativeTasksByParentID(ctx context.Context, parentID string) ([]database.GetCollaborativeTasksByParentIDRow, error)
	GetPublicTasks(ctx context.Context) ([]database.GetPublicTasksRow, error)
	GetParentTasks(ctx context.Context) ([]database.GetParentTasksRow, error)
	UpdateTask(ctx context.Context, task Task) (database.Task, error)
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

// GetTaskByID retrieves a task by its ID from the database
func (s *svc) GetTaskByID(ctx context.Context, id string) (database.Task, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.Task{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	taskID := pgtype.UUID{}
	err = taskID.Scan(id)
	if err != nil {
		return database.Task{}, err
	}
	task, err := qtx.GetTaskByID(ctx, taskID)
	if err != nil {
		return database.Task{}, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return database.Task{}, err
	}
	return task, nil
}

// GetTasksByUserID retrieves all tasks associated with a given user ID from the database
func (s *svc) GetTasksByUserID(ctx context.Context, userID string) ([]database.Task, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	userIDUUID := pgtype.UUID{}
	err = userIDUUID.Scan(userID)
	if err != nil {
		return nil, err
	}
	tasks, err := qtx.GetTasksByUserID(ctx, userIDUUID)
	if err != nil {
		return nil, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetCollaborativeTasksByParentID retrieves all collaborative tasks associated with a given parent task ID or task ID if that user is the owner of the parent ID from the database
func (s *svc) GetCollaborativeTasksByParentID(ctx context.Context, parentID string) ([]database.GetCollaborativeTasksByParentIDRow, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	parentIDUUID := pgtype.UUID{}
	err = parentIDUUID.Scan(parentID)
	if err != nil {
		return nil, err
	}
	tasks, err := qtx.GetCollaborativeTasksByParentID(ctx, parentIDUUID)
	if err != nil {
		return nil, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetPublicTasks retrieves all public tasks from the database
func (s *svc) GetPublicTasks(ctx context.Context) ([]database.GetPublicTasksRow, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	tasks, err := qtx.GetPublicTasks(ctx)
	if err != nil {
		return nil, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetParentTasks retrieves all parent tasks from the database
func (s *svc) GetParentTasks(ctx context.Context) ([]database.GetParentTasksRow, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)
	tasks, err := qtx.GetParentTasks(ctx)
	if err != nil {
		return nil, err
	}
	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateTask updates a task in the database
func (s *svc) UpdateTask(ctx context.Context, task Task) (database.Task, error) {
	// start a transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return database.Task{}, err
	}
	defer tx.Rollback(ctx)
	// create a new Queries instance with the transaction
	qtx := s.repo.WithTx(tx)

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

	updatedTask, err := qtx.UpdateTask(ctx, database.UpdateTaskParams{
		ID:          pgtype.UUID{Bytes: task.ID, Valid: true},
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
	return updatedTask, nil
}
