package middleware

import (
	"fmt"
	"goilerplate/config"
	"mime/multipart"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/gofiber/fiber/v2"
)

type Document struct {
	Timestamp       string              `json:"timestamp"`
	RequestId       interface{}         `json:"request_id"`
	Ip              string              `json:"ip"`
	Method          string              `json:"method"`
	BaseUrl         string              `json:"base_url"`
	Endpoint        string              `json:"endpoint"`
	OriginalUrl     string              `json:"original_url"`
	RequestHeaders  map[string][]string `json:"request_headers"`
	RequestBody     interface{}         `json:"request_body"`
	Status          int                 `json:"status"`
	ResponseHeaders map[string][]string `json:"response_headers"`
	ResponseBody    string              `json:"response_body"`
	Latency         time.Duration       `json:"latency"`
}

func Log(app *fiber.App) {

	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()

		pushToElastic(c)

		return err
	})
}

func pushToElastic(c *fiber.Ctx) {

	document := Document{
		Timestamp:       c.Context().Time().Format("2006-01-02 15:04:05.000"),
		RequestId:       c.Locals("requestid"),
		Ip:              c.IP(),
		Method:          c.Method(),
		BaseUrl:         c.BaseURL(),
		Endpoint:        c.Path(),
		OriginalUrl:     c.BaseURL() + c.OriginalURL(),
		RequestHeaders:  c.GetReqHeaders(),
		RequestBody:     getRequestBody(c.Request()),
		Status:          c.Response().StatusCode(),
		ResponseHeaders: c.GetRespHeaders(),
		ResponseBody:    string(c.Response().Body()),
	}

	latency := time.Since(c.Context().Time())
	document.Latency = latency

	es := config.GetElasticConnection()
	_, err := es.Index("api-log", esutil.NewJSONReader(&document))
	if err != nil {
		fmt.Println("error store log to elastic", err)
		return
	}
}

func getRequestBody(request *fiber.Request) map[string]interface{} {

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
	setFields(&payload, form)
	setFiles(&payload, form)

	return map[string]interface{}{
		"Content-Type": contentType,
		"Payload":      payload.String(),
	}
}

func setFields(payload *strings.Builder, form *multipart.Form) {
	for key, values := range form.Value {
		for _, value := range values {
			appendFields(payload, key, value)
		}
	}
}

func setFiles(payload *strings.Builder, form *multipart.Form) {
	for key, files := range form.File {
		for _, file := range files {
			appendFiles(payload, key, file)
		}
	}

}

func appendFields(payload *strings.Builder, key string, value string) {
	fmt.Fprintf(payload, "%s = %s\n", key, value)
}

func appendFiles(payload *strings.Builder, key string, file *multipart.FileHeader) {
	fmt.Fprintf(payload, "%s = [filename=%s size=%dbytes mimeType=%s]\n",
		key, file.Filename, file.Size, file.Header.Get("Content-Type"))
}
