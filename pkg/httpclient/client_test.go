package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goilerplate/pkg/logger"

	"github.com/stretchr/testify/assert"
)

func TestLoggingRoundTripper(t *testing.T) {
	// Initialize logger
	logger.NewSlog(nil)

	// Setup a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success", "data": {"id": 1}}`))
	}))
	defer server.Close()

	// Create client with our custom transport
	client := NewClient(5 * time.Second)

	// Make a request
	payload := map[string]string{"key": "value"}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", server.URL+"/test?q=1", io.NopCloser(bytes.NewBuffer(bodyBytes)))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom-Header", "test-val")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var jsonResp map[string]interface{}
	err = json.Unmarshal(respBody, &jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, "success", jsonResp["message"])
}
