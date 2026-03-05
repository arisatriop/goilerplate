package bar

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/bar"
	"goilerplate/internal/domain/transaction"
)

// ApplicationService handles multi-domain orchestration
type ApplicationService interface {
	CreateSomething(ctx context.Context, exp *Exp) error
}

type applicationService struct {
	txManager transaction.Transaction
	barUC bar.Usecase

	barRepo bar.Repository
}

func NewApplicationService(
	txManager transaction.Transaction,
	barUC bar.Usecase,
	barRepo bar.Repository,
) ApplicationService {
	return &applicationService{
		txManager:   txManager,
		barUC:   barUC,
		barRepo: barRepo,
	}
}

func (s *applicationService) CreateSomething(ctx context.Context, exp *Exp) error {
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		barRepoWithTx := s.barRepo.WithTx(txCtx)

		_, err := barRepoWithTx.CreateBar(txCtx, exp.Bar)
		if err != nil {
			return fmt.Errorf("failed to create bar: %w", err)
		}

		return nil
	})
}
