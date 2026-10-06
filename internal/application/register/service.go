package register

import (
	"context"
	"fmt"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/role"
	"goilerplate/internal/domain/transaction"
	"goilerplate/internal/domain/user"
	"goilerplate/internal/domain/userrole"
	"goilerplate/pkg/auditctx"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/password"
)

type ApplicationService interface {
	Register(ctx context.Context, register *Register) error
}

// VerificationSender starts email verification for a new account. auth.Usecase satisfies it.
type VerificationSender interface {
	SendEmailVerification(ctx context.Context, email string, origin auth.RequestOrigin) error
}

type applicationService struct {
	txManager      transaction.Transaction
	userRepo       user.Repository
	roleRepo       role.Repository
	userRoleRepo   userrole.Repository
	passwordPolicy password.Policy
	verification   VerificationSender
}

func NewApplicationService(
	txManager transaction.Transaction,
	userRepo user.Repository,
	roleRepo role.Repository,
	userRoleRepo userrole.Repository,
	passwordPolicy password.Policy,
	verification VerificationSender,
) ApplicationService {
	return &applicationService{
		txManager:      txManager,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		passwordPolicy: passwordPolicy,
		verification:   verification,
	}
}

func (s *applicationService) Register(ctx context.Context, register *Register) error {
	// Checking the policy before hashing means bcrypt never silently truncates an over-long
	// password at 72 bytes.
	if err := s.passwordPolicy.Validate(register.Password); err != nil {
		return err
	}

	if err := s.checkExistingEmail(ctx, register.User.Email); err != nil {
		return fmt.Errorf("failed to register new user: %w", err)
	}

	// Hash outside the transaction: bcrypt at cost 12 takes ~250ms, and holding a database
	// connection open for that is wasted contention.
	if err := register.User.SetPassword(register.Password); err != nil {
		return fmt.Errorf("failed to register new user: %w", err)
	}

	role, err := s.roleRepo.GetRoleBySlug(ctx, role.OwnerRoleSlug)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	err = s.txManager.Do(ctx, func(txCtx context.Context) error {
		txCtx = auditctx.WithAuditInfo(txCtx, "system", "system")
		txUserRepo := s.userRepo.WithTx(txCtx)
		txUserRoleRepo := s.userRoleRepo.WithTx(txCtx)

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
	if err != nil {
		return err
	}

	s.sendVerification(ctx, register)
	return nil
}

// sendVerification emails the first verification code, after the account has committed: a use
// case must not run inside the transaction (CLAUDE.md), and a code for a user whose insert
// rolled back would point at nothing. Nil when email is disabled.
//
// A failure is logged, not returned. The account exists either way, and the user can ask for a
// code again; failing the registration would invite them to retry into "already registered".
func (s *applicationService) sendVerification(ctx context.Context, register *Register) {
	if s.verification == nil {
		return
	}
	if err := s.verification.SendEmailVerification(ctx, register.User.Email, register.Origin); err != nil {
		logger.Error(ctx, fmt.Errorf("sending verification email after registration: %w", err))
	}
}

func (s *applicationService) checkExistingEmail(ctx context.Context, email string) error {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return user.ErrEmailAlreadyRegistered
	}
	return nil
}
