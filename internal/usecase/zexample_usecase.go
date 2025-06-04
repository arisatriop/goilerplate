package usecase

import (
	"context"
	"fmt"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/repository"

	"github.com/sirupsen/logrus"
)

type IExampleUsecase interface {
	Create(ctx context.Context, req *model.ExampleCreateRequest) error
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

func (u *ExampleUsecase) Create(ctx context.Context, req *model.ExampleCreateRequest) error {

	err := u.ExampleRepository.Create(u.DB.GDB, nil)
	if err != nil {
		return fmt.Errorf("failed to create example: %w", err)
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
