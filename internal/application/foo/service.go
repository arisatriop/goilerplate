package foo

import (
	"context"
	"goilerplate/internal/domain/bar"
	"goilerplate/internal/domain/foo"
	"goilerplate/internal/domain/transaction"
)

type ApplicationService interface {
	Foo(ctx context.Context) error
}

type applicationService struct {
	txManager transaction.Transaction
	fooUC     foo.Usecase
	fooRepo   foo.Repository
	barRepo   bar.Repository
}

func NewApplicationService(
	tx transaction.Transaction,
	fooUC foo.Usecase,
	fooRepo foo.Repository,
	barRepo bar.Repository,
) ApplicationService {
	return &applicationService{
		txManager: tx,
		fooUC:     fooUC,
		fooRepo:   fooRepo,
		barRepo:   barRepo,
	}
}

func (s *applicationService) Foo(ctx context.Context) error {
	// Implement application logic here

	// Example: Using transaction manager
	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		txFooRepo := s.fooRepo.WithTx(txCtx)
		txBarRepo := s.barRepo.WithTx(txCtx)

		// Use txFooRepo and txBarRepo for database operations within the transaction
		_ = txFooRepo
		_ = txBarRepo

		// Business logic using fooUC, fooRepo, barRepo

		return nil
	})
}
