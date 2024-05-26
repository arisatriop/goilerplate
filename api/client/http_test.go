package client

import (
	"fmt"
	"goilerplate/app/entity"
	"goilerplate/app/logging"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

func TestDo(t *testing.T) {
	t.Run("Test post", func(t *testing.T) {

		var method = "POST"
		var baseUrl = "http://localhost:8000"
		var endpoint = "/test"
		var header map[string]interface{}
		var payload map[string]interface{}

		http := NewHttp()
		_, err := http.Perform(method, baseUrl, endpoint, header, payload)
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
		var payload map[string]interface{}

		http := NewHttp()
		_, err := http.Perform(method, baseUrl, endpoint, header, payload)
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
		var payload map[string]interface{}

		http := NewHttp()
		_, err := http.Perform(method, baseUrl, endpoint, header, payload)
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
		var payload map[string]interface{}

		http := NewHttp()
		_, err := http.Perform(method, baseUrl, endpoint, header, payload)
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
		var payload map[string]interface{}

		http := NewHttp()
		_, err := http.Perform(method, baseUrl, endpoint, header, payload)
		if err != nil {
			t.Errorf("Test Do failed: %s", err)
		}

		if !http.Successful() {
			t.Error("Test Do failed: http return non-200 status code")
		}
	})
}

func TestCurlLog(t *testing.T) {
	t.Run("Test curl log", func(t *testing.T) {

		var method = "GET"
		var baseUrl = "http://localhost:8000"
		var endpoint = "/test"

		header := map[string]interface{}{
			"Authorization": "Bearer token",
			"Content-Type":  "application/json",
		}

		payload := map[string]interface{}{
			"key": "value",
		}

		http := NewHttp()
		result, err := http.Perform(method, baseUrl, endpoint, header, payload)
		if err := curlLog(result); err != nil {
			t.Errorf("error log to elastic: %v", err)
		}

		if err != nil {
			t.Errorf("Test CurlLog failed: %s", err)
		}
	})
}

func curlLog(result *entity.HttpClient) error {

	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost" + ":" + "9200"},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("unable to connect to elasticsearch: %v\n", err)
	}

	curl := logging.NewCurlLog()
	document := curl.GetDocument(result)

	return store(es, document)

}

func store(client *elasticsearch.Client, doc *logging.CurlDocument) error {
	res, err := client.Index("curl-log", esutil.NewJSONReader(doc))
	if err != nil {
		return fmt.Errorf("error %v", err)
	}
	defer res.Body.Close()

	return nil
}
