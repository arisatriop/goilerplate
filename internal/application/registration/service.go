package registration

import (
	"context"
	"fmt"
	"goilerplate/config"
	"goilerplate/internal/domain/plan"
	"goilerplate/internal/domain/plantype"
	"goilerplate/internal/domain/role"
	"goilerplate/internal/domain/store"
	"goilerplate/internal/domain/subscription"
	"goilerplate/internal/domain/transaction"
	"goilerplate/internal/domain/user"
	"goilerplate/internal/domain/userrole"
	"goilerplate/pkg/utils"
	"net/http"

	"github.com/google/uuid"

	auditctx "goilerplate/internal/infrastructure/context"
)

type ApplicationService interface {
	RegisterNewStore(ctx context.Context, registration *Registration) error
}

type applicationService struct {
	cfg              *config.Config
	txManager        transaction.Transaction
	planUc           plan.Usecase
	userRepo         user.Repository
	roleRepo         role.Repository
	userRoleRepo     userrole.Repository
	storeRepo        store.Repository
	plantTypeRepo    plantype.Repository
	planRepo         plan.Repository
	subscriptionRepo subscription.Repository
}

func NewApplicationService(
	cfg *config.Config,
	txManager transaction.Transaction,
	planUC plan.Usecase,
	userRepo user.Repository,
	roleRepo role.Repository,
	userRoleRepo userrole.Repository,
	storeRepo store.Repository,
	plantTypeRepo plantype.Repository,
	planRepo plan.Repository,
	subscriptionRepo subscription.Repository,
) ApplicationService {
	return &applicationService{
		cfg:              cfg,
		txManager:        txManager,
		planUc:           planUC,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		userRoleRepo:     userRoleRepo,
		storeRepo:        storeRepo,
		plantTypeRepo:    plantTypeRepo,
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *applicationService) RegisterNewStore(ctx context.Context, registration *Registration) error {

	if err := s.checkExistingEmail(ctx, registration.User.Email); err != nil {
		return fmt.Errorf("failed to register new store: %w", err)
	}

	role, err := s.roleRepo.GetRoleBySlug(ctx, role.OwnerRoleSlug)
	if err != nil {
		return fmt.Errorf("failed to get role: %v", err)
	}

	plan, err := s.planUc.GetBasicPlan(ctx)
	if err != nil {
		return fmt.Errorf("failed to get basic plan: %v", err)
	}

	return s.txManager.Do(ctx, func(txCtx context.Context) error {
		txCtx = auditctx.WithAuditInfo(txCtx, "system", "system")

		txUserRepo := s.userRepo.WithTx(txCtx)
		txUserRoleRepo := s.userRoleRepo.WithTx(txCtx)
		txStoreRepo := s.storeRepo.WithTx(txCtx)
		txSubscriptionRepo := s.subscriptionRepo.WithTx(txCtx)

		registration.User.HashPassword()
		createdUser, err := txUserRepo.CreateUser(txCtx, registration.User)
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

		createdStore, err := s.createStore(txCtx, txStoreRepo, createdUser.ID, registration.Store)
		if err != nil {
			return fmt.Errorf("failed to create store: %w", err)
		}

		_, err = txSubscriptionRepo.CreateSubscription(txCtx, &subscription.Subscription{
			StoreID:  createdStore.ID,
			PlanID:   plan.ID,
			Price:    plan.Price,
			Status:   "Active",
			IsActive: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
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

func (s *applicationService) createStore(ctx context.Context, storeRepo store.Repository, userID uuid.UUID, store *store.Store) (*store.Store, error) {

	store.ID = uuid.New()
	store.UserID = userID
	store.IsActive = true
	store.GenerateWebURL(s.cfg.Crypto.EncryptionKey, s.cfg.Tenant.BaseURL)

	createdStore, err := storeRepo.CreateStore(ctx, store)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}
	return createdStore, nil
}
