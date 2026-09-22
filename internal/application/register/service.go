package register

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/role"
	"goilerplate/internal/domain/transaction"
	"goilerplate/internal/domain/user"
	"goilerplate/internal/domain/userrole"
	"goilerplate/pkg/auditctx"
	"goilerplate/pkg/password"
	"goilerplate/pkg/utils"
	"net/http"
)

type ApplicationService interface {
	Register(ctx context.Context, register *Register) error
}

type applicationService struct {
	txManager      transaction.Transaction
	userRepo       user.Repository
	roleRepo       role.Repository
	userRoleRepo   userrole.Repository
	passwordPolicy password.Policy
}

func NewApplicationService(
	txManager transaction.Transaction,
	userRepo user.Repository,
	roleRepo role.Repository,
	userRoleRepo userrole.Repository,
	passwordPolicy password.Policy,
) ApplicationService {
	return &applicationService{
		txManager:      txManager,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		passwordPolicy: passwordPolicy,
	}
}

func (s *applicationService) Register(ctx context.Context, register *Register) error {
	// The plaintext password still sits in PasswordHash at this point; HashPassword replaces it
	// below. Checking the policy first means bcrypt never silently truncates an over-long one.
	if err := s.passwordPolicy.Validate(register.User.PasswordHash); err != nil {
		return err
	}

	if err := s.checkExistingEmail(ctx, register.User.Email); err != nil {
		return fmt.Errorf("failed to register new user: %w", err)
	}

	role, err := s.roleRepo.GetRoleBySlug(ctx, role.OwnerRoleSlug)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
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
		return utils.ClientErr(http.StatusBadRequest, "email is already registered")
	}
	return nil
}
