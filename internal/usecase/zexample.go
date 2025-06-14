package usecase

import (
	"context"
	"fmt"
	"goilerplate/internal/config"
	"goilerplate/internal/model/zexample"
	"goilerplate/internal/repository"
	"goilerplate/pkg/helper"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ExampleUsecase interface {
	Get(ctx context.Context, id uuid.UUID) (*zexample.GetResponse, error)
	GetAll(ctx context.Context, req *zexample.GetRequest) ([]zexample.GetAllResponse, int64, error)
	Create(ctx context.Context, req *zexample.CreateRequest) error
	Update(ctx context.Context, id uuid.UUID, req *zexample.UpdateRequest) error
	Delete(ctx context.Context, id uuid.UUID, req *zexample.DeleteRequest) error
}

type exampleUsecase struct {
	Log               *logrus.Logger
	DB                *config.DB
	ExampleRepository repository.ExampleRepository
}

func NewExampleUsecase(log *logrus.Logger, db *config.DB, exampleRepo repository.ExampleRepository) ExampleUsecase {
	return &exampleUsecase{
		Log:               log,
		DB:                db,
		ExampleRepository: exampleRepo,
	}
}

func (u *exampleUsecase) Get(ctx context.Context, id uuid.UUID) (*zexample.GetResponse, error) {
	example, err := u.ExampleRepository.GetByID(u.DB.GDB.WithContext(ctx), id)
	if err != nil {
		return nil, err
	}

	if example == nil {
		return nil, helper.Error(http.StatusNotFound, fmt.Sprintf("example with ID %s not found", id))
	}

	return zexample.ToGet(example), nil
}

func (u *exampleUsecase) GetAll(ctx context.Context, req *zexample.GetRequest) ([]zexample.GetAllResponse, int64, error) {

	result, total, err := u.ExampleRepository.GetAll(u.DB.GDB.WithContext(ctx), req)
	if err != nil {
		return nil, 0, err
	}

	var response []zexample.GetAllResponse
	for _, example := range result {
		response = append(response, *zexample.ToGetAll(&example))
	}

	return response, total, nil

}

func (u *exampleUsecase) Create(ctx context.Context, req *zexample.CreateRequest) error {

	example, err := req.ToCreate()
	if err != nil {
		return err
	}

	return u.ExampleRepository.Create(u.DB.GDB.WithContext(ctx), example)
}

func (u *exampleUsecase) Update(ctx context.Context, id uuid.UUID, req *zexample.UpdateRequest) error {

	example, err := u.ExampleRepository.GetByID(u.DB.GDB.WithContext(ctx), id)
	if err != nil {
		return err
	}

	if example == nil {
		return helper.Error(http.StatusNotFound, fmt.Sprintf("example with ID %s not found", id))
	}

	err = req.ToUpdate(example)
	if err != nil {
		return err
	}

	return u.ExampleRepository.Update(u.DB.GDB.WithContext(ctx), example)
}

func (u *exampleUsecase) Delete(ctx context.Context, id uuid.UUID, req *zexample.DeleteRequest) error {
	db := u.DB.GDB.WithContext(ctx)

	example, err := u.ExampleRepository.GetByID(db, id)
	if err != nil {
		return err
	}

	if example == nil {
		return helper.Error(http.StatusNotFound, fmt.Sprintf("example with ID %s not found", id))
	}

	req.ToDelete(example)

	return u.ExampleRepository.Delete(db, example)
}
