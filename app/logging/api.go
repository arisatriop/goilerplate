package logging

import (
	"encoding/json"
	"fmt"
	"goilerplate/config"
	"mime/multipart"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/gofiber/fiber/v2"
)

type ApiDocument struct {
	Timestamp          time.Time              `json:"timestamp"`
	RequestId          interface{}            `json:"request_id"`
	Ip                 string                 `json:"ip"`
	Method             string                 `json:"method"`
	BaseUrl            string                 `json:"base_url"`
	Endpoint           string                 `json:"endpoint"`
	OriginalUrl        string                 `json:"original_url"`
	RequestHeaders     map[string][]string    `json:"request_headers"`
	RequestBody        interface{}            `json:"request_body"`
	Status             int                    `json:"status"`
	ResponseHeaders    map[string][]string    `json:"response_headers"`
	ResponseBody       map[string]interface{} `json:"response_body"`
	ResponseBodyString string                 `json:"response_body_string"`
	Latency            string                 `json:"latency"`
}

type ApiLog struct{}

func (log *ApiLog) Store(c *fiber.Ctx) error {

	document := log.GetDocument(c)

	es := config.GetElasticConnection()
	res, err := es.Index("api-log", esutil.NewJSONReader(&document))
	if err != nil {
		return fmt.Errorf("error %v", err)
	}
	defer res.Body.Close()

	return nil

}

func (log *ApiLog) GetDocument(c *fiber.Ctx) *ApiDocument {
	return &ApiDocument{
		Timestamp:          c.Context().Time(),
		RequestId:          c.Locals("requestid"),
		Ip:                 c.IP(),
		Method:             c.Method(),
		BaseUrl:            c.BaseURL(),
		Endpoint:           c.Path(),
		OriginalUrl:        c.BaseURL() + c.OriginalURL(),
		RequestHeaders:     c.GetReqHeaders(),
		RequestBody:        log.GetRequestBody(c.Request()),
		Status:             c.Response().StatusCode(),
		ResponseHeaders:    c.GetRespHeaders(),
		ResponseBody:       log.GetResponseBody(c.Response().Body()),
		ResponseBodyString: string(c.Response().Body()),
		Latency:            time.Since(c.Context().Time()).String(),
	}
}

func (log *ApiLog) GetRequestBody(request *fiber.Request) map[string]interface{} {

	contentType := string(request.Header.ContentType())

	if contentType == "application/json" {
		return map[string]interface{}{
			"Content-Type": contentType,
			"Payload":      string(request.Body()),
		}
	}

	form, err := request.MultipartForm()
	if err != nil {
		return map[string]interface{}{
			"Content-Type": contentType,
			"Payload":      "failed to parse multipart form",
		}
	}

	var payload strings.Builder
	log.SetFields(&payload, form)
	log.SetFiles(&payload, form)

	return map[string]interface{}{
		"Content-Type": contentType,
		"Payload":      payload.String(),
	}
}

func (log *ApiLog) GetResponseBody(body []byte) map[string]interface{} {
	// Define a map to store the unmarshaled data
	var response map[string]interface{}

	// Unmarshal the JSON data into the map
	err := json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return response
}

func (log *ApiLog) SetFields(payload *strings.Builder, form *multipart.Form) {
	for key, values := range form.Value {
		for _, value := range values {
			log.AppendFields(payload, key, value)
		}
	}
}

func (log *ApiLog) SetFiles(payload *strings.Builder, form *multipart.Form) {
	for key, files := range form.File {
		for _, file := range files {
			log.AppendFiles(payload, key, file)
		}
	}

}

func (log *ApiLog) AppendFields(payload *strings.Builder, key string, value string) {
	fmt.Fprintf(payload, "%s = %s\n", key, value)
}

func (log *ApiLog) AppendFiles(payload *strings.Builder, key string, file *multipart.FileHeader) {
	fmt.Fprintf(payload, "%s = [filename=%s size=%dbytes mimeType=%s]\n",
		key, file.Filename, file.Size, file.Header.Get("Content-Type"))
}
