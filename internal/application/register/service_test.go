package register_test

import (
	"context"
	"testing"

	"goilerplate/internal/application/register"
	"goilerplate/internal/domain/role"
	"goilerplate/internal/domain/user"
	"goilerplate/internal/domain/userrole"
	"goilerplate/pkg/password"
	"goilerplate/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The stubs below stand in for the three repositories and the transaction. A real database
// cannot answer the question these tests ask — "was CreateUser reached at all?" — and the
// orchestration under test is exactly the layer that decides that.

type stubUserRepo struct {
	created  *user.User
	byEmail  *user.User
	findCall int
}

func (r *stubUserRepo) WithTx(context.Context) user.Repository { return r }

func (r *stubUserRepo) FindByEmail(context.Context, string) (*user.User, error) {
	r.findCall++
	return r.byEmail, nil
}

func (r *stubUserRepo) CreateUser(_ context.Context, u *user.User) (*user.User, error) {
	r.created = u
	stored := *u
	stored.ID = uuid.New()
	return &stored, nil
}

type stubRoleRepo struct{ role *role.Role }

func (r *stubRoleRepo) WithTx(context.Context) role.Repository { return r }

func (r *stubRoleRepo) GetRoleBySlug(context.Context, string) (*role.Role, error) {
	return r.role, nil
}

type stubUserRoleRepo struct{ created *userrole.UserRole }

func (r *stubUserRoleRepo) WithTx(context.Context) userrole.Repository { return r }

func (r *stubUserRoleRepo) CreateUserRole(_ context.Context, ur *userrole.UserRole) error {
	r.created = ur
	return nil
}

// inlineTx runs the body without a real transaction. The transaction semantics belong to the
// repository suite; what matters here is whether the body runs at all.
type inlineTx struct{ entered bool }

func (t *inlineTx) Do(ctx context.Context, fn func(context.Context) error) error {
	t.entered = true
	return fn(ctx)
}

func newService(t *testing.T) (register.ApplicationService, *stubUserRepo, *stubUserRoleRepo, *inlineTx) {
	t.Helper()

	userRepo := &stubUserRepo{}
	userRoleRepo := &stubUserRoleRepo{}
	tx := &inlineTx{}
	roleRepo := &stubRoleRepo{role: &role.Role{ID: uuid.New(), Slug: role.OwnerRoleSlug}}

	svc := register.NewApplicationService(tx, userRepo, roleRepo, userRoleRepo, password.NewPolicy(nil))
	return svc, userRepo, userRoleRepo, tx
}

func TestRegister_PersistsAHashNeverThePlaintext(t *testing.T) {
	// Arrange
	svc, userRepo, userRoleRepo, _ := newService(t)
	input := &register.Register{
		User:     &user.User{Name: "Ada", Email: "ada@example.com"},
		Password: "a-perfectly-fine-password",
	}

	// Act
	err := svc.Register(context.Background(), input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, userRepo.created)
	assert.NotEqual(t, "a-perfectly-fine-password", userRepo.created.PasswordHash,
		"the plaintext must never reach the repository")
	assert.NoError(t, utils.CheckPassword("a-perfectly-fine-password", userRepo.created.PasswordHash))
	assert.NotNil(t, userRoleRepo.created, "the owner role is assigned in the same unit of work")
}

// A password the policy rejects must stop before anything is read or written. Hashing is the
// expensive step and creating the user is the irreversible one.
func TestRegister_RejectedPasswordNeverReachesTheRepositories(t *testing.T) {
	// Arrange
	svc, userRepo, userRoleRepo, tx := newService(t)
	input := &register.Register{
		User:     &user.User{Name: "Ada", Email: "ada@example.com"},
		Password: "short",
	}

	// Act
	err := svc.Register(context.Background(), input)

	// Assert
	require.Error(t, err)
	assert.Zero(t, userRepo.findCall, "no lookup should happen for input we already refused")
	assert.Nil(t, userRepo.created)
	assert.Nil(t, userRoleRepo.created)
	assert.False(t, tx.entered, "no transaction should be opened")
	assert.Empty(t, input.User.PasswordHash)
}

func TestRegister_DuplicateEmailIsRefusedBeforeHashing(t *testing.T) {
	// Arrange
	svc, userRepo, userRoleRepo, tx := newService(t)
	userRepo.byEmail = &user.User{ID: uuid.New(), Email: "ada@example.com"}
	input := &register.Register{
		User:     &user.User{Name: "Ada", Email: "ada@example.com"},
		Password: "a-perfectly-fine-password",
	}

	// Act
	err := svc.Register(context.Background(), input)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrEmailAlreadyRegistered)
	assert.Nil(t, userRepo.created)
	assert.Nil(t, userRoleRepo.created)
	assert.False(t, tx.entered)
	assert.Empty(t, input.User.PasswordHash, "no point paying for bcrypt on a request already lost")
}
