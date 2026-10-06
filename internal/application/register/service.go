package register

import (
	"context"
	"errors"
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

// AccountMailer is the email registration sends. auth.Usecase satisfies it.
type AccountMailer interface {
	// SendEmailVerification starts verification for a new account.
	SendEmailVerification(ctx context.Context, email string, origin auth.RequestOrigin) error
	// SendAccountExistsNotice tells the owner of a taken address that it was registered again.
	SendAccountExistsNotice(ctx context.Context, email string) error
}

type applicationService struct {
	txManager      transaction.Transaction
	userRepo       user.Repository
	roleRepo       role.Repository
	userRoleRepo   userrole.Repository
	passwordPolicy password.Policy
	mailer         AccountMailer
}

func NewApplicationService(
	txManager transaction.Transaction,
	userRepo user.Repository,
	roleRepo role.Repository,
	userRoleRepo userrole.Repository,
	passwordPolicy password.Policy,
	mailer AccountMailer,
) ApplicationService {
	return &applicationService{
		txManager:      txManager,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		passwordPolicy: passwordPolicy,
		mailer:         mailer,
	}
}

// Register creates an account. What it says about an address that is already registered
// depends on whether email is configured (mailer is non-nil):
//
//   - With email, it says nothing. A taken address gets the same nil as a new one, and its owner
//     is emailed instead, so the endpoint cannot be used to find out which addresses have
//     accounts. Both paths hash the password, so they also take the same time.
//   - Without email, there is no one to tell but the caller, so it returns
//     user.ErrEmailAlreadyRegistered (409), and skips bcrypt for a request it has already lost.
func (s *applicationService) Register(ctx context.Context, register *Register) error {
	// Checking the policy before hashing means bcrypt never silently truncates an over-long
	// password at 72 bytes.
	if err := s.passwordPolicy.Validate(register.Password); err != nil {
		return err
	}

	if s.mailer != nil {
		// Before the lookup: a taken address must cost what a new one does.
		if err := register.User.SetPassword(register.Password); err != nil {
			return fmt.Errorf("failed to register new user: %w", err)
		}
	}

	err := s.checkExistingEmail(ctx, register.User.Email)
	if errors.Is(err, user.ErrEmailAlreadyRegistered) && s.mailer != nil {
		s.notifyExisting(ctx, register.User.Email)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to register new user: %w", err)
	}

	if s.mailer == nil {
		// Hash outside the transaction: bcrypt at cost 12 takes ~250ms, and holding a database
		// connection open for that is wasted contention.
		if err := register.User.SetPassword(register.Password); err != nil {
			return fmt.Errorf("failed to register new user: %w", err)
		}
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
	// A concurrent registration of the same address won the insert. It is the same case as the
	// lookup finding it, and answered the same way.
	if errors.Is(err, user.ErrEmailAlreadyRegistered) && s.mailer != nil {
		s.notifyExisting(ctx, register.User.Email)
		return nil
	}
	if err != nil {
		return err
	}

	s.sendVerification(ctx, register)
	return nil
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

// sendVerification emails the first verification code, after the account has committed: a use
// case must not run inside the transaction (CLAUDE.md), and a code for a user whose insert
// rolled back would point at nothing.
//
// A failure is logged, not returned. The account exists either way, and the user can ask for a
// code again; failing the registration would invite them to retry into "already registered".
func (s *applicationService) sendVerification(ctx context.Context, register *Register) {
	if s.mailer == nil {
		return
	}
	if err := s.mailer.SendEmailVerification(ctx, register.User.Email, register.Origin); err != nil {
		logger.Error(ctx, fmt.Errorf("sending verification email after registration: %w", err))
	}
}

// notifyExisting emails the owner of a taken address. A failure is logged, not returned: an error
// for a taken address and success for a new one is the difference this path exists to hide.
func (s *applicationService) notifyExisting(ctx context.Context, email string) {
	if err := s.mailer.SendAccountExistsNotice(ctx, email); err != nil {
		logger.Error(ctx, fmt.Errorf("sending account exists notice: %w", err))
	}
}
