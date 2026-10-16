package service

import (
	"context"
	"go-codeexplorer/dtos"
	"go-codeexplorer/models"

	"github.com/jackc/pgx/v5"
)

type TodoEntryRepository interface {
	Add(ctx context.Context, todo models.TodoEntry) (models.TodoEntry, error)
	GetByID(ctx context.Context, id int64) (models.TodoEntry, error)
	List(ctx context.Context) ([]models.TodoEntry, error)
	MarkDone(ctx context.Context, id int64) error
}

type TransactionalManager interface {
	WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type TodoService struct {
	repository TodoEntryRepository
	txManager  TransactionalManager
}

func NewTodoService(entryRepository TodoEntryRepository, manager TransactionalManager) *TodoService {
	return &TodoService{entryRepository, manager}
}

func (s *TodoService) Add(ctx context.Context, request dtos.CreateNewTodoEntryRequest) (dtos.CreateNewTodoEntryResponse, error) {
	var add models.TodoEntry
	var err error
	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		add, err = s.repository.Add(ctx, models.TodoEntry{Description: request.Description})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return dtos.CreateNewTodoEntryResponse{}, err
	}
	return dtos.CreateNewTodoEntryResponse{
		Id:          add.Id,
		CreatedAt:   add.CreatedAt,
		Description: add.Description,
		Done:        add.Done,
	}, nil
}

func (s *TodoService) List(ctx context.Context) ([]dtos.TodoEntryResponse, error) {
	var response []dtos.TodoEntryResponse
	err := s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		list, err := s.repository.List(ctx)
		if err != nil {
			return err
		}

		response = Map(list, func(it models.TodoEntry) dtos.TodoEntryResponse {
			return dtos.TodoEntryResponse{
				Id:          it.Id,
				CreatedAt:   it.CreatedAt,
				Description: it.Description,
				Done:        it.Done,
			}
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *TodoService) GetByID(ctx context.Context, id int64) (dtos.TodoEntryResponse, error) {
	var add models.TodoEntry
	var err error
	err = s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		add, err = s.repository.GetByID(ctx, id)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return dtos.TodoEntryResponse{}, err
	}

	return dtos.TodoEntryResponse{
		Id:          add.Id,
		CreatedAt:   add.CreatedAt,
		Description: add.Description,
		Done:        add.Done,
	}, nil
}

func (s *TodoService) MarkDone(ctx context.Context, id int64) error {
	return s.txManager.WithTx(ctx, func(tx pgx.Tx) error {
		return s.repository.MarkDone(ctx, id)
	})
}

func Map[T, R any](s []T, f func(T) R) []R {
	res := make([]R, len(s))
	for i, v := range s {
		res[i] = f(v)
	}
	return res
}
