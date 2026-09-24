package bar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"goilerplate/pkg/utils"

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

func assertClientStatus(t *testing.T, err error, status int) {
	t.Helper()
	var clientErr *utils.ClientError
	require.ErrorAs(t, err, &clientErr)
	assert.Equal(t, status, clientErr.Code)
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
	assertClientStatus(t, err, http.StatusConflict)
}

func TestUsecase_Create_InvalidInputNeverReachesTheRepository(t *testing.T) {
	repo := &fakeRepo{}

	_, err := NewUseCase(repo).Create(context.Background(), &Bar{Code: "NOPE", Bar: "thing"})

	assertClientStatus(t, err, http.StatusBadRequest)
	assert.Nil(t, repo.created)
}

func TestUsecase_Create_RepositoryFailureIsNotAClientError(t *testing.T) {
	repo := &fakeRepo{writeErr: errors.New("connection reset")}

	_, err := NewUseCase(repo).Create(context.Background(), &Bar{Code: "EXP-1", Bar: "thing"})

	require.Error(t, err)
	var clientErr *utils.ClientError
	assert.False(t, errors.As(err, &clientErr))
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

	assertClientStatus(t, err, http.StatusNotFound)
}

func TestUsecase_BulkCreate(t *testing.T) {
	tests := []struct {
		name       string
		entities   []*Bar
		writeErr   error
		wantStatus int
		wantWrite  bool
	}{
		{"all new", []*Bar{{Code: "EXP-1", Bar: "a"}, {Code: "EXP-2", Bar: "b"}}, nil, 0, true},
		{"same code twice in the request, differing only in case", []*Bar{{Code: "EXP-1", Bar: "a"}, {Code: " exp-1", Bar: "b"}},
			nil, http.StatusBadRequest, false},
		{"clash with a stored bar", []*Bar{{Code: "EXP-1", Bar: "a"}}, ErrCodeAlreadyExists, http.StatusConflict, true},
		{"one invalid entity", []*Bar{{Code: "EXP-1", Bar: "a"}, {Code: "BAD", Bar: "b"}}, nil, http.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{writeErr: tt.writeErr}

			err := NewUseCase(repo).BulkCreate(context.Background(), tt.entities)

			if tt.wantStatus == 0 {
				require.NoError(t, err)
			} else {
				assertClientStatus(t, err, tt.wantStatus)
			}
			assert.Equal(t, tt.wantWrite, repo.bulk != nil, "whether the batch reached the repository")
		})
	}
}
