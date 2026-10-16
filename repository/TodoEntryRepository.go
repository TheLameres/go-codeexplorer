package repository

import (
	"context"
	"errors"
	"fmt"
	"go-codeexplorer/models"

	"github.com/jackc/pgx/v5"
)

type TodoEntryRepository struct {
	db DBTX
}

func NewTodoEntryRepository(db DBTX) *TodoEntryRepository {
	return &TodoEntryRepository{db: db}
}

func (r *TodoEntryRepository) Add(ctx context.Context, todo models.TodoEntry) (models.TodoEntry, error) {
	const q = `
		insert into todos (description)
		values ($1)
		returning id, created_at, description, done`

	var res models.TodoEntry
	err := r.db.QueryRow(ctx, q, todo.Description).
		Scan(&res.Id, &res.CreatedAt, &res.Description, &res.Done)
	if err != nil {
		return models.TodoEntry{}, fmt.Errorf("insert todo: %w", err)
	}
	return res, nil
}

func (r *TodoEntryRepository) GetByID(ctx context.Context, id int64) (models.TodoEntry, error) {
	const q = `select id, created_at, description, done from todos where id = $1`

	var res models.TodoEntry
	err := r.db.QueryRow(ctx, q, id).
		Scan(&res.Id, &res.CreatedAt, &res.Description, &res.Done)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TodoEntry{}, models.ErrNotFound
	}
	if err != nil {
		return models.TodoEntry{}, fmt.Errorf("get todo %d: %w", id, err)
	}
	return res, nil
}

func (r *TodoEntryRepository) List(ctx context.Context) ([]models.TodoEntry, error) {
	const q = `select id, created_at, description, done from todos order by id`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	todos, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.TodoEntry, error) {
		var t models.TodoEntry
		err := row.Scan(&t.Id, &t.CreatedAt, &t.Description, &t.Done)
		return t, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect todos: %w", err)
	}
	return todos, nil
}

func (r *TodoEntryRepository) MarkDone(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `update todos set done = true where id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark done: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}
