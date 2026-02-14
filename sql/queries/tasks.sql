-- name: CreateTask :one
INSERT INTO tasks (id, created_at, updated_at, user_id, title, end_date, description, priority, tag, state, parent_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: CreateTaskEditors :one
INSERT INTO task_editors (task_id, editor_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetTaskByID :one
SELECT * FROM tasks WHERE id = $1;  

-- name: GetTaskEditorsByTaskID :many
SELECT editor_id FROM task_editors WHERE task_id = $1;

-- name: GetTasksByUserID :many
SELECT * FROM tasks WHERE user_id = $1 ORDER BY created_at ASC;

-- name: GetCollaborativeTasksByParentID :many
SELECT u.email, u.username, t.*
FROM tasks t
JOIN users u ON t.user_id = u.id
WHERE t.parent_id = $1 OR t.id = $1
ORDER BY t.created_at ASC;
