package example

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/example"
	"goilerplate/internal/domain/transaction"
)

// ApplicationService handles multi-domain orchestration
type ApplicationService interface {
	CreateSomething(ctx context.Context, exp *Exp) error
}

type applicationService struct {
	txManager transaction.Transaction
	exampleUC example.Usecase

	exampleRepo example.Repository
}

func NewApplicationService(
	txManager transaction.Transaction,
	exampleUC example.Usecase,
	exampleRepo example.Repository,
) ApplicationService {
	return &applicationService{
		txManager:   txManager,
		exampleUC:   exampleUC,
		exampleRepo: exampleRepo,
	}
}

func (s *applicationService) CreateSomething(ctx context.Context, exp *Exp) error {
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		exampleRepoWithTx := s.exampleRepo.WithTx(txCtx)

		_, err := exampleRepoWithTx.CreateExample(txCtx, exp.Example)
		if err != nil {
			return fmt.Errorf("failed to create example: %w", err)
		}

		return nil
	})
}
