package usecase

import (
	"context"
	"fmt"
	"golang-clean-architecture/internal/config"
	"golang-clean-architecture/internal/model"
	"golang-clean-architecture/internal/repository"
)

type ExampleUsecase struct {
	DB                *config.DB
	ExampleRepository repository.IExampleRepository
}

type IExampleUsecase interface {
	FindAll(ctx context.Context, req *model.ExampleGetRequest) ([]model.ExampleListReponse, error)
}

func NewExampleUsecase(db *config.DB, exampleRepo repository.IExampleRepository) IExampleUsecase {
	return &ExampleUsecase{
		DB:                db,
		ExampleRepository: exampleRepo,
	}
}

func (u *ExampleUsecase) FindAll(ctx context.Context, req *model.ExampleGetRequest) ([]model.ExampleListReponse, error) {
	examples, err := u.ExampleRepository.FindAll(ctx, u.DB.GDB, req)
	if err != nil {
		return nil, fmt.Errorf("failed to find examples: %w", err)
	}

	result := []model.ExampleListReponse{}
	for _, example := range examples {
		result = append(result, model.ToExampleListResponse(&example))
	}

	return result, nil
}
