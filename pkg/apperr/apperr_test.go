package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"goilerplate/pkg/apperr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errThingNotFound = apperr.New(apperr.NotFound, "thing_not_found", "Thing not found")

func TestError_MatchesItsSentinelThroughWrappingAndCause(t *testing.T) {
	cause := errors.New("no rows")
	err := fmt.Errorf("getting thing: %w", errThingNotFound.WithCause(cause))

	assert.ErrorIs(t, err, errThingNotFound)
	assert.ErrorIs(t, err, cause, "the cause stays reachable for logging")
	assert.Equal(t, "getting thing: Thing not found", err.Error())

	appErr, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.NotFound, appErr.Kind)
	assert.Equal(t, "thing_not_found", appErr.Code)
}

func TestError_DifferentCodeOrKindDoesNotMatch(t *testing.T) {
	assert.NotErrorIs(t, apperr.New(apperr.NotFound, "other_not_found", "x"), errThingNotFound)
	assert.NotErrorIs(t, apperr.New(apperr.Conflict, "thing_not_found", "x"), errThingNotFound)
	assert.NotErrorIs(t, errors.New("Thing not found"), errThingNotFound, "never matched by message")
}

func TestError_WithCauseLeavesTheSentinelUntouched(t *testing.T) {
	_ = errThingNotFound.WithCause(errors.New("x"))

	assert.NoError(t, errors.Unwrap(errThingNotFound))
}

func TestAs_NonClientError(t *testing.T) {
	_, ok := apperr.As(errors.New("connection reset"))

	assert.False(t, ok)
}
