package entity

import (
	"bytes"
)

type HttpClient struct {
	Request  *Request
	Response *Response
	Latency  string
}

type Request struct {
	Method      string
	BaseURL     string
	Endpoint    string
	OriginalURL string
	Header      map[string]interface{}
	Payload     *bytes.Buffer
	Payloads    map[string]interface{}
}

type Response struct {
	StatusCode int
	Header     map[string]interface{}
	Body       map[string]interface{}
}
