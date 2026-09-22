package presenter_test

import (
	"testing"

	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/domain/bar"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToBarResponse(t *testing.T) {
	entity := &bar.Bar{ID: "abc", Code: "EXP001", Bar: "first"}

	dto := presenter.ToBarResponse(entity)

	require.NotNil(t, dto)
	assert.Equal(t, "abc", dto.ID)
	assert.Equal(t, "EXP001", dto.Code)
	assert.Equal(t, "first", dto.Bar)
}

func TestToBarListResponse(t *testing.T) {
	entities := []*bar.Bar{
		{ID: "1", Code: "EXP001", Bar: "first"},
		{ID: "2", Code: "EXP002", Bar: "second"},
	}

	dtos := presenter.ToBarListResponse(entities)

	require.Len(t, dtos, 2)
	assert.Equal(t, "1", dtos[0].ID)
	assert.Equal(t, "second", dtos[1].Bar)
}

// An empty result must produce an empty slice, not nil. response.Paginated normalises nil
// anyway, but a presenter returning nil would make every other caller handle it too.
func TestToBarListResponse_EmptyInputIsAnEmptySlice(t *testing.T) {
	dtos := presenter.ToBarListResponse(nil)

	assert.NotNil(t, dtos)
	assert.Empty(t, dtos)
}
