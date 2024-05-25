package client

import (
	"testing"
)

func TestDo(t *testing.T) {
	t.Run("Test post", func(t *testing.T) {

		var method = "POST"
		var baseUrl = "http://localhost:8000"
		var endpoint = "/test"
		var header map[string]interface{}
		var payload interface{}

		http := NewHttp()
		err := http.Perform(method, baseUrl, endpoint, header, payload)
		if err != nil {
			t.Errorf("Test Do failed: %s", err)
		}

		if !http.Successful() {
			t.Error("Test Do failed: http return non-200 status code")
		}
	})

	t.Run("Test put", func(t *testing.T) {

		var method = "PUT"
		var baseUrl = "http://localhost:8000"
		var endpoint = "/test"
		var header map[string]interface{}
		var payload interface{}

		http := NewHttp()
		err := http.Perform(method, baseUrl, endpoint, header, payload)
		if err != nil {
			t.Errorf("Test Do failed: %s", err)
		}

		if !http.Successful() {
			t.Error("Test Do failed: http return non-200 status code")
		}
	})

	t.Run("Test patch", func(t *testing.T) {

		var method = "PATCH"
		var baseUrl = "http://localhost:8000"
		var endpoint = "/test"
		var header map[string]interface{}
		var payload interface{}

		http := NewHttp()
		err := http.Perform(method, baseUrl, endpoint, header, payload)
		if err != nil {
			t.Errorf("Test Do failed: %s", err)
		}

		if !http.Successful() {
			t.Error("Test Do failed: http return non-200 status code")
		}
	})

	t.Run("Test delete", func(t *testing.T) {

		var method = "DELETE"
		var baseUrl = "http://localhost:8000"
		var endpoint = "/test"
		var header map[string]interface{}
		var payload interface{}

		http := NewHttp()
		err := http.Perform(method, baseUrl, endpoint, header, payload)
		if err != nil {
			t.Errorf("Test Do failed: %s", err)
		}

		if !http.Successful() {
			t.Error("Test Do failed: http return non-200 status code")
		}
	})

	t.Run("Test get", func(t *testing.T) {

		var method = "GET"
		var baseUrl = "http://localhost:8000"
		var endpoint = "/test"
		var header map[string]interface{}
		var payload interface{}

		http := NewHttp()
		err := http.Perform(method, baseUrl, endpoint, header, payload)
		if err != nil {
			t.Errorf("Test Do failed: %s", err)
		}

		if !http.Successful() {
			t.Error("Test Do failed: http return non-200 status code")
		}
	})
}
