package example

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/transaction"
	"goilerplate/internal/domain/zexample"
)

// ApplicationService handles multi-domain orchestration
type ApplicationService interface {
	CreateSomething(ctx context.Context, exp *Exp) error
}

type applicationService struct {
	txManager transaction.Transaction
	exampleUC zexample.Usecase

	exampleRepo zexample.Repository
}

func NewApplicationService(
	txManager transaction.Transaction,
	exampleUC zexample.Usecase,
	exampleRepo zexample.Repository,
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
