package logging

import (
	"encoding/json"
	"fmt"
	"goilerplate/app/entity"
	"goilerplate/config"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type CurlDocument struct {
	Timestamp            time.Time              `json:"timestamp"`
	RequestId            interface{}            `json:"request_id"`
	Method               string                 `json:"method"`
	BaseUrl              string                 `json:"base_url"`
	Endpoint             string                 `json:"endpoint"`
	OriginalUrl          string                 `json:"original_url"`
	RequestHeaders       map[string]interface{} `json:"request_header"`
	RequestHeaderString  string                 `json:"request_header_string"`
	RequestBody          map[string]interface{} `json:"request_body"`
	RequestBodyString    string                 `json:"request_body_string"`
	Status               int                    `json:"status"`
	ResponseHeaders      map[string]interface{} `json:"response_header"`
	ResponseHeaderString string                 `json:"response_header_string"`
	ResponseBody         map[string]interface{} `json:"response_body"`
	ResponseBodyString   string                 `json:"response_body_string"`
	Latency              string                 `json:"latency"`
}

type CurlLog interface {
	Store(result *entity.HttpClient) error
	StoreToFile(document *CurlDocument) error
	GetDocument(result *entity.HttpClient) *CurlDocument
}

type CurlLogImpl struct{}

func NewCurlLog() CurlLog {
	return &CurlLogImpl{}
}

func (log *CurlLogImpl) Store(result *entity.HttpClient) error {
	document := log.GetDocument(result)

	go log.StoreToFile(document)

	es := config.GetElasticConnection()
	res, err := es.Index("curl-log", esutil.NewJSONReader(&document))
	if err != nil {
		// return fmt.Errorf("error %v", err)
		return nil // always return nil
	}
	defer res.Body.Close()

	return nil // always return nil
}

func (log *CurlLogImpl) GetDocument(result *entity.HttpClient) *CurlDocument {
	return &CurlDocument{
		Timestamp:            time.Now(),
		Method:               result.Request.Method,
		BaseUrl:              result.Request.BaseURL,
		Endpoint:             result.Request.Endpoint,
		OriginalUrl:          result.Request.OriginalURL,
		RequestHeaders:       result.Request.Header,
		RequestHeaderString:  toString(result.Request.Header),
		RequestBody:          result.Request.Payloads,
		RequestBodyString:    toString(result.Request.Payloads),
		Status:               result.Response.StatusCode,
		ResponseHeaders:      result.Response.Header,
		ResponseHeaderString: toString(result.Response.Header),
		ResponseBody:         result.Response.Body,
		ResponseBodyString:   toString(result.Response.Body),
		Latency:              result.Latency,
	}
}

func (log *CurlLogImpl) StoreToFile(document *CurlDocument) error {
	file, _ := os.OpenFile("./logs/curl.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	file.WriteString(fmt.Sprintf("%s | %s | %s | %d\n", document.Timestamp.Format("2006-01-02 15:04:05.000"), document.Method, document.Endpoint, document.Status))

	return nil // always return nil
}

func toString(m map[string]interface{}) string {
	jsonString, err := json.Marshal(m)
	if err != nil {
		return "error on parsing json"
	}

	return string(jsonString)
}
