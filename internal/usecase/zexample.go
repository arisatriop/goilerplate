package usecase

import (
	"context"
	"fmt"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/helper"
	"golang-clean-architecture/internal/model/zexample"
	"golang-clean-architecture/internal/repository"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type IExampleUsecase interface {
	FindByID(ctx context.Context, id uuid.UUID) (*zexample.GetResponse, error)
	Create(ctx context.Context, req *zexample.CreateRequest) error
	Update(ctx context.Context, id uuid.UUID, req *zexample.UpdateRequest) error
	Delete(ctx context.Context, id uuid.UUID, req *zexample.DeleteRequest) error
	// FindAll(ctx context.Context, req *model.ExampleGetRequest) ([]model.ExampleListReponse, error)
}

type ExampleUsecase struct {
	Log               *logrus.Logger
	DB                *config.DB
	ExampleRepository repository.IExampleRepository
}

func NewExampleUsecase(log *logrus.Logger, db *config.DB, exampleRepo repository.IExampleRepository) IExampleUsecase {
	return &ExampleUsecase{
		Log:               log,
		DB:                db,
		ExampleRepository: exampleRepo,
	}
}

func (u *ExampleUsecase) FindByID(ctx context.Context, id uuid.UUID) (*zexample.GetResponse, error) {
	example, err := u.ExampleRepository.FindByID(u.DB.GDB, id)
	if err != nil {
		return nil, err
	}

	if example == nil {
		return nil, helper.Error(http.StatusNotFound, fmt.Sprintf("example with ID %s not found", id))
	}

	return zexample.ToGetResponse(example), nil
}

func (u *ExampleUsecase) Create(ctx context.Context, req *zexample.CreateRequest) error {

	example, err := req.ToCreate()
	if err != nil {
		return err
	}

	err = u.ExampleRepository.Create(u.DB.GDB, example)
	if err != nil {
		return err
	}
	return nil
}

func (u *ExampleUsecase) Update(ctx context.Context, id uuid.UUID, req *zexample.UpdateRequest) error {

	example, err := u.ExampleRepository.FindByID(u.DB.GDB, id)
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

	err = u.ExampleRepository.Update(u.DB.GDB, example)
	if err != nil {
		return err
	}

	return nil
}

func (u *ExampleUsecase) Delete(ctx context.Context, id uuid.UUID, req *zexample.DeleteRequest) error {
	example, err := u.ExampleRepository.FindByID(u.DB.GDB, id)
	if err != nil {
		return err
	}

	if example == nil {
		return helper.Error(http.StatusNotFound, fmt.Sprintf("example with ID %s not found", id))
	}

	req.ToDelete(example)

	err = u.ExampleRepository.Update(u.DB.GDB, example)
	if err != nil {
		return err
	}

	return nil
}

// func (u *ExampleUsecase) FindAll(ctx context.Context, req *model.ExampleGetRequest) ([]model.ExampleListReponse, error) {
// 	examples, err := u.ExampleRepository.FindAll(ctx, u.DB.GDB, req)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to find examples: %w", err)
// 	}

// 	result := []model.ExampleListReponse{}
// 	for _, example := range examples {
// 		result = append(result, model.ToExampleListResponse(&example))
// 	}

// 	return result, nil
// }
