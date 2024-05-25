package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpImpl struct{ Http *Http }

type Http struct {
	Request  *Request
	Response *Response
}

type Request struct {
	Method      string
	BaseURL     string
	Endpoint    string
	OriginalURL string
	Header      map[string]interface{}
	Payload     *bytes.Buffer
	RawPayload  interface{}
}

type Response struct {
	StatusCode int
	Header     map[string]interface{}
	Body       map[string]interface{}
}

type IHttp interface {
	Perform(method string, baseURL string, endpoint string, header map[string]interface{}, payload interface{}) error
	Post() error
	Put() error
	Patch() error
	Get() error
	Delete() error
	Do(request *http.Request) error

	SetBaseURL(baseURL string)
	SetEndpoint(endpoint string)
	SetOriginalURL()
	SetRequestHeader(data map[string]interface{})
	SetRequestPayload(data interface{}) error
	SetResponseHeader(response *http.Response)
	SetResponseBody(response *http.Response) error

	Successful() bool
	Result() *Http
}

func NewHttp() IHttp {
	return &HttpImpl{
		Http: &Http{
			Request:  &Request{},
			Response: &Response{},
		},
	}
}

func (h *HttpImpl) Perform(method string, baseURL string, endpoint string, header map[string]interface{}, payload interface{}) error {

	h.SetBaseURL(baseURL)
	h.SetEndpoint(endpoint)
	h.SetOriginalURL()
	h.SetRequestHeader(header)

	if err := h.SetRequestPayload(payload); err != nil {
		return fmt.Errorf("perform: %v", err)
	}

	switch method {
	case "POST":
		return h.Post()
	case "PUT":
		return h.Put()
	case "PATCH":
		return h.Patch()
	case "GET":
		return h.Get()
	case "DELETE":
		return h.Delete()
	}

	return fmt.Errorf("method %s not supported", method)
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

func (h *HttpImpl) SetRequestPayload(data interface{}) error {

	var payload bytes.Buffer

	if data != nil {
		payload = *new(bytes.Buffer)
		if err := json.NewEncoder(&payload).Encode(data); err != nil {
			return fmt.Errorf("set payload: %v", err)
		}
	}

	h.Http.Request.RawPayload = data
	h.Http.Request.Payload = &payload
	return nil
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

func (h *HttpImpl) Result() *Http {
	return h.Http
}
