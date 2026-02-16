package expapp

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/transaction"
	"goilerplate/internal/domain/zexample"
	"goilerplate/internal/domain/zexample2"
)

// ApplicationService handles multi-domain orchestration
type ApplicationService interface {
	CreateSomething(ctx context.Context, exp *Exp) error
}

type applicationService struct {
	txManager transaction.Transaction
	exampleUC zexample.Usecase

	exampleRepo  zexample.Repository
	example2Repo zexample2.Repository
}

func NewApplicationService(
	txManager transaction.Transaction,
	exampleUC zexample.Usecase,
	exampleRepo zexample.Repository,
	example2Repo zexample2.Repository,
) ApplicationService {
	return &applicationService{
		txManager:    txManager,
		exampleUC:    exampleUC,
		exampleRepo:  exampleRepo,
		example2Repo: example2Repo,
	}
}

func (s *applicationService) CreateSomething(ctx context.Context, exp *Exp) error {
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		exampleRepoWithTx := s.exampleRepo.WithTx(txCtx)
		example2RepoWithTx := s.example2Repo.WithTx(txCtx)

		_, err := exampleRepoWithTx.CreateExample(txCtx, exp.Example)
		if err != nil {
			return fmt.Errorf("failed to create example: %w", err)
		}

		for _, example2 := range exp.Example2 {
			_, err = example2RepoWithTx.CreateExample(txCtx, example2)
			if err != nil {
				return fmt.Errorf("failed to create example2: %w", err)
			}
		}

		return nil
	})
}
