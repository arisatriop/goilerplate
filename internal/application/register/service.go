package register

import (
	"context"
	"fmt"
	"goilerplate/config"
	"goilerplate/internal/domain/role"
	"goilerplate/internal/domain/transaction"
	"goilerplate/internal/domain/user"
	"goilerplate/internal/domain/userrole"
	"goilerplate/pkg/utils"
	"net/http"

	auditctx "goilerplate/internal/infrastructure/context"
)

type ApplicationService interface {
	Register(ctx context.Context, regiter *Register) error
}

type applicationService struct {
	cfg          *config.Config
	txManager    transaction.Transaction
	userRepo     user.Repository
	roleRepo     role.Repository
	userRoleRepo userrole.Repository
}

func NewApplicationService(
	cfg *config.Config,
	txManager transaction.Transaction,
	userRepo user.Repository,
	roleRepo role.Repository,
	userRoleRepo userrole.Repository,
) ApplicationService {
	return &applicationService{
		cfg:          cfg,
		txManager:    txManager,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
	}
}

func (s *applicationService) Register(ctx context.Context, register *Register) error {
	if err := s.checkExistingEmail(ctx, register.User.Email); err != nil {
		return fmt.Errorf("failed to register new user: %w", err)
	}

	role, err := s.roleRepo.GetRoleBySlug(ctx, role.OwnerRoleSlug)
	if err != nil {
		return fmt.Errorf("failed to get role: %v", err)
	}

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		txCtx = auditctx.WithAuditInfo(txCtx, "system", "system")
		txUserRepo := s.userRepo.WithTx(txCtx)
		txUserRoleRepo := s.userRoleRepo.WithTx(txCtx)

		register.User.HashPassword()
		createdUser, err := txUserRepo.CreateUser(txCtx, register.User)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		txCtx = auditctx.WithAuditInfo(txCtx, createdUser.ID.String(), createdUser.Name)
		if err := txUserRoleRepo.CreateUserRole(txCtx, &userrole.UserRole{
			UserID: createdUser.ID,
			RoleID: role.ID,
		}); err != nil {
			return fmt.Errorf("failed to assign role to user: %w", err)
		}

		return nil
	})
}

func (s *applicationService) checkExistingEmail(ctx context.Context, email string) error {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return utils.Error(http.StatusBadRequest, "email is already registered")
	}
	return nil
}
