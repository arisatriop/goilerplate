package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goilerplate/app/entity"
	"net/http"
)

type HttpImpl struct{ Http *entity.HttpClient }

type IHttp interface {
	Perform(method string, baseURL string, endpoint string, header map[string]interface{}, payload map[string]interface{}) (*entity.HttpClient, error)
	Post() error
	Put() error
	Patch() error
	Get() error
	Delete() error
	Do(request *http.Request) error

	SetMethod(method string)
	SetBaseURL(baseURL string)
	SetEndpoint(endpoint string)
	SetOriginalURL()
	SetRequestHeader(data map[string]interface{})
	SetRequestPayload(data map[string]interface{}) error
	SetRequestPayloads(data map[string]interface{})
	SetResponseHeader(response *http.Response)
	SetResponseBody(response *http.Response) error

	Successful() bool
	Result() *entity.HttpClient
}

func NewHttp() IHttp {
	return &HttpImpl{
		Http: &entity.HttpClient{
			Request:  &entity.Request{},
			Response: &entity.Response{},
		},
	}
}

func (h *HttpImpl) Perform(method string, baseURL string, endpoint string,
	header map[string]interface{}, payload map[string]interface{}) (*entity.HttpClient, error) {

	h.SetMethod(method)
	h.SetBaseURL(baseURL)
	h.SetEndpoint(endpoint)
	h.SetOriginalURL()
	h.SetRequestHeader(header)
	h.SetRequestPayload(payload)

	var err error

	switch method {
	case "POST":
		err = h.Post()
	case "PUT":
		err = h.Put()
	case "PATCH":
		err = h.Patch()
	case "GET":
		err = h.Get()
	case "DELETE":
		err = h.Delete()
	default:
		return nil, fmt.Errorf("method %s not supported", method)
	}

	return h.Http, err
}

func (h *HttpImpl) Post() error {

	request, err := http.NewRequest("POST", h.Http.Request.OriginalURL, h.Http.Request.Payload)
	if err != nil {
		return fmt.Errorf("post: %v", err)
	}

	if err := h.Do(request); err != nil {
		return fmt.Errorf("post: %v", err)
	}

	return nil
}

func (h *HttpImpl) Put() error {
	request, err := http.NewRequest("PUT", h.Http.Request.OriginalURL, h.Http.Request.Payload)
	if err != nil {
		return fmt.Errorf("put: %v", err)
	}

	if err := h.Do(request); err != nil {
		return fmt.Errorf("put: %v", err)
	}

	return nil
}

func (h *HttpImpl) Patch() error {
	request, err := http.NewRequest("PATCH", h.Http.Request.OriginalURL, h.Http.Request.Payload)
	if err != nil {
		return fmt.Errorf("patch: %v", err)
	}

	if err := h.Do(request); err != nil {
		return fmt.Errorf("patch: %v", err)
	}

	return nil
}

func (h *HttpImpl) Get() error {
	request, err := http.NewRequest("GET", h.Http.Request.OriginalURL, h.Http.Request.Payload)
	if err != nil {
		return fmt.Errorf("get: %v", err)
	}

	if err := h.Do(request); err != nil {
		return fmt.Errorf("get: %v", err)
	}

	return nil
}

func (h *HttpImpl) Delete() error {
	request, err := http.NewRequest("DELETE", h.Http.Request.OriginalURL, h.Http.Request.Payload)
	if err != nil {
		return fmt.Errorf("delete: %v", err)
	}

	if err := h.Do(request); err != nil {
		return fmt.Errorf("delete: %v", err)
	}

	return nil
}

func (h *HttpImpl) Do(request *http.Request) error {
	client := new(http.Client)

	response, err := client.Do(request)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("do: %v", err)
	}

	h.SetStatusCode(response)
	h.SetResponseHeader(response)

	if err := h.SetResponseBody(response); err != nil {
		return fmt.Errorf("do: %v", err)
	}

	return nil
}

func (h *HttpImpl) SetRequestHeader(data map[string]interface{}) {
	h.Http.Request.Header = data
}

func (h *HttpImpl) SetRequestPayload(data map[string]interface{}) error {
	var payload bytes.Buffer

	if data != nil {
		payload = *new(bytes.Buffer)
		if err := json.NewEncoder(&payload).Encode(data); err != nil {
			return fmt.Errorf("set payload: %v", err)
		}
	}

	h.Http.Request.Payload = &payload
	h.Http.Request.Payloads = data
	return nil
}

func (h *HttpImpl) SetRequestPayloads(data map[string]interface{}) {
	h.Http.Request.Payloads = data
}

func (h *HttpImpl) SetStatusCode(response *http.Response) {
	h.Http.Response.StatusCode = response.StatusCode
}

func (h *HttpImpl) SetResponseHeader(response *http.Response) {

	header := make(map[string]interface{})
	for key, values := range response.Header {
		header[key] = values
	}
	h.Http.Response.Header = header
}

func (h *HttpImpl) SetResponseBody(response *http.Response) error {

	body := make(map[string]interface{})
	if response.Header.Get("Content-Type") == "application/json" {
		err := json.NewDecoder(response.Body).Decode(&body)
		if err != nil {
			return fmt.Errorf("set response body: %v", err)
		}
	}

	h.Http.Response.Body = body
	return nil
}

func (h *HttpImpl) SetMethod(method string) {
	h.Http.Request.Method = method
}

func (h *HttpImpl) SetBaseURL(baseURL string) {
	h.Http.Request.BaseURL = baseURL
}

func (h *HttpImpl) SetEndpoint(endpoint string) {
	h.Http.Request.Endpoint = endpoint
}

func (h *HttpImpl) SetOriginalURL() {
	h.Http.Request.OriginalURL = h.Http.Request.BaseURL + h.Http.Request.Endpoint
}

func (h *HttpImpl) Successful() bool {
	return h.Http.Response.StatusCode >= 200 && h.Http.Response.StatusCode < 300
}

func (h *HttpImpl) Result() *entity.HttpClient {
	return h.Http
}
