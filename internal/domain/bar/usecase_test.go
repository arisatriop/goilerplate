package bar

import (
	"context"
	"errors"
	"fmt"
	"goilerplate/pkg/apperr"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRepo records writes; methods the tests do not reach panic through the nil embedded
// interface.
type fakeRepo struct {
	Repository
	created  *Bar
	updated  *Bar
	bulk     []*Bar
	writeErr error
}

func (r *fakeRepo) CreateBar(_ context.Context, entity *Bar) (*Bar, error) {
	r.created = entity
	return entity, r.writeErr
}

func (r *fakeRepo) UpdateBar(_ context.Context, entity *Bar) error {
	r.updated = entity
	return r.writeErr
}

func (r *fakeRepo) BulkCreate(_ context.Context, entities []*Bar) error {
	r.bulk = entities
	return r.writeErr
}

func assertKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	appErr, ok := apperr.As(err)
	require.True(t, ok, "want a client error, got %v", err)
	assert.Equal(t, kind, appErr.Kind)
}

func TestUsecase_Create_StoresTheNormalisedForm(t *testing.T) {
	// Arrange
	repo := &fakeRepo{}

	// Act
	_, err := NewUseCase(repo).Create(context.Background(), &Bar{Code: "  exp-1 ", Bar: " thing "})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "EXP-1", repo.created.Code)
	assert.Equal(t, "thing", repo.created.Bar)
}

// The database decides uniqueness; the use case must pass its verdict through as a 409 rather
// than burying it under a 500.
func TestUsecase_Create_ConflictFromTheRepositoryIs409(t *testing.T) {
	repo := &fakeRepo{writeErr: ErrCodeAlreadyExists}

	_, err := NewUseCase(repo).Create(context.Background(), &Bar{Code: "EXP-1", Bar: "thing"})

	assert.ErrorIs(t, err, ErrCodeAlreadyExists)
	assertKind(t, err, apperr.Conflict)
}

func TestUsecase_Create_InvalidInputNeverReachesTheRepository(t *testing.T) {
	repo := &fakeRepo{}

	_, err := NewUseCase(repo).Create(context.Background(), &Bar{Code: "NOPE", Bar: "thing"})

	assertKind(t, err, apperr.Invalid)
	assert.Nil(t, repo.created)
}

func TestUsecase_Create_RepositoryFailureIsNotAClientError(t *testing.T) {
	repo := &fakeRepo{writeErr: errors.New("connection reset")}

	_, err := NewUseCase(repo).Create(context.Background(), &Bar{Code: "EXP-1", Bar: "thing"})

	require.Error(t, err)
	_, isClientErr := apperr.As(err)
	assert.False(t, isClientErr)
}

// The old code compared the stored code with the request's before normalising, so re-saving a
// bar with its own code in lower case looked like a clash with itself and answered 409.
func TestUsecase_Update_ChangingOnlyTheCaseOfItsOwnCodeIsNotAConflict(t *testing.T) {
	repo := &fakeRepo{}

	updated, err := NewUseCase(repo).Update(context.Background(), &Bar{ID: "b1", Code: "exp-1", Bar: "thing"})

	require.NoError(t, err)
	assert.Equal(t, "EXP-1", repo.updated.Code)
	assert.Equal(t, "EXP-1", updated.Code)
}

func TestUsecase_Update_NotFoundPassesThrough(t *testing.T) {
	repo := &fakeRepo{writeErr: fmt.Errorf("wrapped: %w", ErrNotFound)}

	_, err := NewUseCase(repo).Update(context.Background(), &Bar{ID: "b1", Code: "EXP-1", Bar: "thing"})

	assertKind(t, err, apperr.NotFound)
}

func TestUsecase_BulkCreate(t *testing.T) {
	tests := []struct {
		name      string
		entities  []*Bar
		writeErr  error
		wantKind  apperr.Kind
		wantWrite bool
	}{
		{"all new", []*Bar{{Code: "EXP-1", Bar: "a"}, {Code: "EXP-2", Bar: "b"}}, nil, 0, true},
		{"same code twice in the request, differing only in case", []*Bar{{Code: "EXP-1", Bar: "a"}, {Code: " exp-1", Bar: "b"}},
			nil, apperr.Invalid, false},
		{"clash with a stored bar", []*Bar{{Code: "EXP-1", Bar: "a"}}, ErrCodeAlreadyExists, apperr.Conflict, true},
		{"one invalid entity", []*Bar{{Code: "EXP-1", Bar: "a"}, {Code: "BAD", Bar: "b"}}, nil, apperr.Invalid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{writeErr: tt.writeErr}

			err := NewUseCase(repo).BulkCreate(context.Background(), tt.entities)

			if tt.wantKind == 0 {
				require.NoError(t, err)
			} else {
				assertKind(t, err, tt.wantKind)
			}
			assert.Equal(t, tt.wantWrite, repo.bulk != nil, "whether the batch reached the repository")
		})
	}
}
