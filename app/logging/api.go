package logging

import (
	"encoding/json"
	"fmt"
	"goilerplate/config"
	"mime/multipart"
	"os"
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

type ApiLog interface {
	Store(c *fiber.Ctx) error
	StoreFile(document *ApiDocument) error
	GetDocument(c *fiber.Ctx) *ApiDocument
	GetRequestBody(request *fiber.Request) map[string]interface{}
	GetResponseBody(body []byte) map[string]interface{}
	SetFields(payload *strings.Builder, form *multipart.Form)
	SetFiles(payload *strings.Builder, form *multipart.Form)
	AppendFields(payload *strings.Builder, key string, value string)
	AppendFiles(payload *strings.Builder, key string, file *multipart.FileHeader)
}

type ApiLogImpl struct{}

func NewApiLog() ApiLog {
	return &ApiLogImpl{}
}

func (log *ApiLogImpl) Store(c *fiber.Ctx) error {

	document := log.GetDocument(c)

	go log.StoreFile(document)

	es := config.GetElasticConnection()
	res, err := es.Index("api-log", esutil.NewJSONReader(&document))
	if err != nil {
		// return fmt.Errorf("error %v", err)
		return nil // always return nil
	}
	defer res.Body.Close()

	return nil

}

func (log *ApiLogImpl) StoreFile(document *ApiDocument) error {
	file, err := os.OpenFile("./logs/api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error open file: ", err.Error())
		return nil
	}

	if _, err := file.WriteString(fmt.Sprintf("\n%s | %s | %s | %s | %s | %d | %s\n--Header:\n%v\n--Payload:\n%v\n--Response:\n%v\n",
		document.Timestamp.Format("2006-01-02 15:04:05.000"),
		document.RequestId,
		document.RequestHeaders["X-User"][0],
		document.Method,
		document.Endpoint,
		document.Status,
		document.Latency,
		document.RequestHeaders,
		document.RequestBody,
		document.ResponseBody,
	)); err != nil {
		fmt.Println("error store log to file: ", err.Error())
		return nil
	}

	return nil //  return alway nil
}

func (log *ApiLogImpl) GetDocument(c *fiber.Ctx) *ApiDocument {
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

func (log *ApiLogImpl) GetRequestBody(request *fiber.Request) map[string]interface{} {

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
			"Payload":      []string{},
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

func (log *ApiLogImpl) GetResponseBody(body []byte) map[string]interface{} {
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

func (log *ApiLogImpl) SetFields(payload *strings.Builder, form *multipart.Form) {
	for key, values := range form.Value {
		for _, value := range values {
			log.AppendFields(payload, key, value)
		}
	}
}

func (log *ApiLogImpl) SetFiles(payload *strings.Builder, form *multipart.Form) {
	for key, files := range form.File {
		for _, file := range files {
			log.AppendFiles(payload, key, file)
		}
	}

}

func (log *ApiLogImpl) AppendFields(payload *strings.Builder, key string, value string) {
	fmt.Fprintf(payload, "%s = %s\n", key, value)
}

func (log *ApiLogImpl) AppendFiles(payload *strings.Builder, key string, file *multipart.FileHeader) {
	fmt.Fprintf(payload, "%s = [filename=%s size=%dbytes mimeType=%s]\n",
		key, file.Filename, file.Size, file.Header.Get("Content-Type"))
}
