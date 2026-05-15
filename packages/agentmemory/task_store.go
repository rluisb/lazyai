package agentmemory

import (
	"context"
	"database/sql"
	"time"
)

// TaskStore persists task lifecycle state.
type TaskStore struct{ db *sql.DB }

func NewTaskStore(db *sql.DB) *TaskStore { return &TaskStore{db: db} }

func (s *TaskStore) CreateTask(ctx context.Context, task Task) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tasks (id, namespace, project_root, task_type, state, current_step, state_json, goal, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.Namespace, task.ProjectRoot, task.TaskType, task.State, task.CurrentStep,
		Redact(task.StateJSON), task.Goal, task.Tags, task.CreatedAt, task.UpdatedAt)
	return err
}

func (s *TaskStore) GetTask(ctx context.Context, id string) (Task, error) {
	return scanTask(s.db.QueryRowContext(ctx, `
		SELECT id, namespace, project_root, task_type, state, COALESCE(current_step, ''), COALESCE(state_json, ''),
		       COALESCE(goal, ''), COALESCE(tags, ''), created_at, updated_at
		FROM tasks WHERE id = ?`, id))
}

func (s *TaskStore) UpdateTaskState(ctx context.Context, id string, state string, currentStep string, stateJSON string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE tasks SET state = ?, current_step = ?, state_json = ?, updated_at = ? WHERE id = ?`,
		state, currentStep, Redact(stateJSON), time.Now().UTC().Format(time.RFC3339Nano), id)
	return err
}

func (s *TaskStore) ListTasks(ctx context.Context, namespace string) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, namespace, project_root, task_type, state, COALESCE(current_step, ''), COALESCE(state_json, ''),
		       COALESCE(goal, ''), COALESCE(tags, ''), created_at, updated_at
		FROM tasks WHERE namespace = ? ORDER BY updated_at DESC, id DESC`, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

type taskScanner interface{ Scan(dest ...any) error }

func scanTask(row taskScanner) (Task, error) {
	var task Task
	err := row.Scan(&task.ID, &task.Namespace, &task.ProjectRoot, &task.TaskType, &task.State,
		&task.CurrentStep, &task.StateJSON, &task.Goal, &task.Tags, &task.CreatedAt, &task.UpdatedAt)
	return task, err
}
