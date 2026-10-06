package register_test

import (
	"context"
	"errors"
	"testing"

	"goilerplate/internal/application/register"
	"goilerplate/internal/domain/auth"
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

	svc := register.NewApplicationService(tx, userRepo, roleRepo, userRoleRepo, password.NewPolicy(nil), nil)
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

// recordingMailer records what registration asked to send.
type recordingMailer struct {
	verified []string
	existing []string
	origin   auth.RequestOrigin
	err      error
}

func (m *recordingMailer) SendEmailVerification(_ context.Context, email string, origin auth.RequestOrigin) error {
	m.verified = append(m.verified, email)
	m.origin = origin
	return m.err
}

func (m *recordingMailer) SendAccountExistsNotice(_ context.Context, email string) error {
	m.existing = append(m.existing, email)
	return m.err
}

// failingUserRoleRepo makes the transaction fail after the user insert.
type failingUserRoleRepo struct{}

func (r *failingUserRoleRepo) WithTx(context.Context) userrole.Repository { return r }

func (r *failingUserRoleRepo) CreateUserRole(context.Context, *userrole.UserRole) error {
	return errors.New("insert failed")
}

// racingUserRepo finds no user, then loses the insert to a concurrent registration.
type racingUserRepo struct{ stubUserRepo }

func (r *racingUserRepo) WithTx(context.Context) user.Repository { return r }

func (r *racingUserRepo) CreateUser(context.Context, *user.User) (*user.User, error) {
	return nil, user.ErrEmailAlreadyRegistered
}

func newServiceWithMailer(t *testing.T, users user.Repository, userRoles userrole.Repository, mailer register.AccountMailer) register.ApplicationService {
	t.Helper()
	roleRepo := &stubRoleRepo{role: &role.Role{ID: uuid.New(), Slug: role.OwnerRoleSlug}}
	if mailer == nil {
		return register.NewApplicationService(&inlineTx{}, users, roleRepo, userRoles, password.NewPolicy(nil), nil)
	}
	return register.NewApplicationService(&inlineTx{}, users, roleRepo, userRoles, password.NewPolicy(nil), mailer)
}

func newInput() *register.Register {
	return &register.Register{
		User:     &user.User{Name: "Ada", Email: "ada@example.com"},
		Password: "a-perfectly-fine-password",
	}
}

func TestRegister_SendsTheFirstVerificationCode(t *testing.T) {
	mailer := &recordingMailer{}
	svc := newServiceWithMailer(t, &stubUserRepo{}, &stubUserRoleRepo{}, mailer)
	input := newInput()
	input.Origin = auth.RequestOrigin{IPAddress: "203.0.113.7", UserAgent: "curl/8"}

	err := svc.Register(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, []string{"ada@example.com"}, mailer.verified)
	assert.Equal(t, input.Origin, mailer.origin)
	assert.Empty(t, mailer.existing)
}

// The account exists once the transaction commits. A code that cannot be sent is the user's to
// ask for again, not a reason to fail a registration that already happened.
func TestRegister_SucceedsWhenTheCodeCannotBeSent(t *testing.T) {
	svc := newServiceWithMailer(t, &stubUserRepo{}, &stubUserRoleRepo{}, &recordingMailer{err: errors.New("database unavailable")})

	assert.NoError(t, svc.Register(context.Background(), newInput()))
}

func TestRegister_NoCodeForARolledBackAccount(t *testing.T) {
	mailer := &recordingMailer{}
	svc := newServiceWithMailer(t, &stubUserRepo{}, &failingUserRoleRepo{}, mailer)

	err := svc.Register(context.Background(), newInput())

	require.Error(t, err)
	assert.Empty(t, mailer.verified, "a code for an account that does not exist points at nothing")
}

// With email on, registration must not say which addresses are taken. A taken address gets the
// same nil as a new one, its owner is emailed instead, and the password is hashed on both paths
// so the response time does not tell them apart either.
func TestRegister_TakenAddressLooksLikeANewOne(t *testing.T) {
	users := &stubUserRepo{byEmail: &user.User{ID: uuid.New(), Email: "ada@example.com"}}
	mailer := &recordingMailer{}
	svc := newServiceWithMailer(t, users, &stubUserRoleRepo{}, mailer)
	input := newInput()

	err := svc.Register(context.Background(), input)

	require.NoError(t, err)
	assert.Nil(t, users.created, "nothing is created")
	assert.Equal(t, []string{"ada@example.com"}, mailer.existing, "the owner is told instead")
	assert.Empty(t, mailer.verified)
	assert.NotEmpty(t, input.User.PasswordHash, "bcrypt runs here too, or timing would give it away")
}

// The same answer when a concurrent registration wins the insert after the lookup.
func TestRegister_LosingTheInsertRaceLooksLikeANewOne(t *testing.T) {
	mailer := &recordingMailer{}
	svc := newServiceWithMailer(t, &racingUserRepo{}, &stubUserRoleRepo{}, mailer)

	err := svc.Register(context.Background(), newInput())

	require.NoError(t, err)
	assert.Equal(t, []string{"ada@example.com"}, mailer.existing)
}

// Without email there is nobody to tell but the caller, so the 409 stays.
func TestRegister_WithoutEmailATakenAddressIsAConflict(t *testing.T) {
	users := &stubUserRepo{byEmail: &user.User{ID: uuid.New(), Email: "ada@example.com"}}
	svc := newServiceWithMailer(t, users, &stubUserRoleRepo{}, nil)

	err := svc.Register(context.Background(), newInput())

	assert.ErrorIs(t, err, user.ErrEmailAlreadyRegistered)
}
