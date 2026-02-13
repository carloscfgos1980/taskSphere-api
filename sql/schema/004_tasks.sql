-- +goose Up
CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    end_date TIMESTAMP NOT NULL,
    description TEXT NOT NULL,
    priority TEXT NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    tag TEXT NOT NULL CHECK (tag IN ('private', 'collaborative')),
    state TEXT NOT NULL CHECK (state IN ('pending', 'in progress', 'done', 'cancelled')),
    parent_id UUID REFERENCES tasks(id)
);

CREATE TABLE task_editors (
    task_id UUID NOT NULL REFERENCES tasks(id),
    editor_id UUID NOT NULL REFERENCES users(id),
    PRIMARY KEY (task_id, editor_id)
);

-- +goose Down
DROP TABLE tasks;