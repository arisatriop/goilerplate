package v1

import (
	"goilerplate/api/request"
	"testing"
)

func TestGetPlaceholder(t *testing.T) {
	t.Run("Test case 1", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Search: "test",
			Limit:  "10",
			Offset: "10",
		}
		search, limit, offset := getPlaceHolder(&payload)
		if search != "$1" {
			t.Errorf("want: $1, got: %s", search)
		}
		if limit != "$2" {
			t.Errorf("want: $2, got: %s", limit)
		}
		if offset != "$3" {
			t.Errorf("want: $3, got: %s", offset)
		}
	})

	t.Run("Test case 2", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Search: "test",
			Limit:  "10",
		}
		search, limit, offset := getPlaceHolder(&payload)
		if search != "$1" {
			t.Errorf("want: $1, got: %s", search)
		}
		if limit != "$2" {
			t.Errorf("want: $2, got: %s", limit)
		}
		if offset != "$0" {
			t.Errorf("want: $0, got: %s", offset)
		}
	})

	t.Run("Test case 3", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Search: "test",
			Offset: "10",
		}
		search, limit, offset := getPlaceHolder(&payload)
		if search != "$1" {
			t.Errorf("want: $1, got: %s", search)
		}
		if limit != "$0" {
			t.Errorf("want: $0, got: %s", limit)
		}
		if offset != "$2" {
			t.Errorf("want: $2, got: %s", offset)
		}
	})

	t.Run("Test case 4", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Search: "test",
		}
		search, limit, offset := getPlaceHolder(&payload)
		if search != "$1" {
			t.Errorf("want: $1, got: %s", search)
		}
		if limit != "$0" {
			t.Errorf("want: $0, got: %s", limit)
		}
		if offset != "$0" {
			t.Errorf("want: $0, got: %s", offset)
		}
	})

	t.Run("Test case 5", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Limit:  "10",
			Offset: "10",
		}
		search, limit, offset := getPlaceHolder(&payload)
		if search != "$0" {
			t.Errorf("want: $0, got: %s", search)
		}
		if limit != "$1" {
			t.Errorf("want: $1, got: %s", limit)
		}
		if offset != "$2" {
			t.Errorf("want: $2, got: %s", offset)
		}
	})

	t.Run("Test case 6", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Limit: "10",
		}
		search, limit, offset := getPlaceHolder(&payload)
		if search != "$0" {
			t.Errorf("want: $0, got: %s", search)
		}
		if limit != "$1" {
			t.Errorf("want: $1, got: %s", limit)
		}
		if offset != "$0" {
			t.Errorf("want: $0, got: %s", offset)
		}
	})

	t.Run("Test case 7", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Offset: "10",
		}
		search, limit, offset := getPlaceHolder(&payload)
		if search != "$0" {
			t.Errorf("want: $0, got: %s", search)
		}
		if limit != "$0" {
			t.Errorf("want: $0, got: %s", limit)
		}
		if offset != "$1" {
			t.Errorf("want: $1, got: %s", offset)
		}
	})

	t.Run("Test case 8", func(t *testing.T) {
		payload := request.ExampleReadPayload{}
		search, limit, offset := getPlaceHolder(&payload)
		if search != "$0" {
			t.Errorf("want: $0, got: %s", search)
		}
		if limit != "$0" {
			t.Errorf("want: $0, got: %s", limit)
		}
		if offset != "$0" {
			t.Errorf("want: $0, got: %s", offset)
		}
	})
}

func TestGetSqlFindAllExample(t *testing.T) {

	t.Run("Test case 1", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Search: "test",
			Limit:  "10",
			Offset: "10",
		}

		got := getSqlFindAllExample(&payload)
		want := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null and (code ilike '%' || $1 || '%' or example ilike '%' || $1 || '%') limit $2 offset $3"

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}

	})

	t.Run("Test case 2", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Search: "test",
			Limit:  "10",
		}

		got := getSqlFindAllExample(&payload)
		want := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null and (code ilike '%' || $1 || '%' or example ilike '%' || $1 || '%') limit $2"

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}

	})

	t.Run("Test case 3", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Search: "test",
			Offset: "10",
		}

		got := getSqlFindAllExample(&payload)
		want := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null and (code ilike '%' || $1 || '%' or example ilike '%' || $1 || '%') offset $2"

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}

	})

	t.Run("Test case 4", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Search: "test",
		}

		got := getSqlFindAllExample(&payload)
		want := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null and (code ilike '%' || $1 || '%' or example ilike '%' || $1 || '%')"

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}

	})

	t.Run("Test case 5", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Limit:  "10",
			Offset: "10",
		}

		got := getSqlFindAllExample(&payload)
		want := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null limit $1 offset $2"

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}

	})

	t.Run("Test case 6", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Limit: "10",
		}

		got := getSqlFindAllExample(&payload)
		want := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null limit $1"

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}

	})

	t.Run("Test case 7", func(t *testing.T) {
		payload := request.ExampleReadPayload{
			Offset: "10",
		}

		got := getSqlFindAllExample(&payload)
		want := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null offset $1"

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}

	})

	t.Run("Test case 8", func(t *testing.T) {
		payload := request.ExampleReadPayload{}

		got := getSqlFindAllExample(&payload)
		want := "select id,code,example,created_at,created_by,updated_at,updated_by,deleted_at,deleted_by,uuid from example where deleted_at is null"

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}

	})

}
