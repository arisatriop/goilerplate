package middleware

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

type Document struct {
	Timestamp       time.Time              `json:"timestamp"`
	RequestId       interface{}            `json:"request_id"`
	Ip              string                 `json:"ip"`
	Method          string                 `json:"method"`
	BaseUrl         string                 `json:"base_url"`
	Endpoint        string                 `json:"endpoint"`
	OriginalUrl     string                 `json:"original_url"`
	RequestHeaders  map[string][]string    `json:"request_headers"`
	RequestBody     interface{}            `json:"request_body"`
	Status          int                    `json:"status"`
	ResponseHeaders map[string][]string    `json:"response_headers"`
	ResponseBody    map[string]interface{} `json:"response_body"`
	Latency         string                 `json:"latency"`
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
		Timestamp:       c.Context().Time(),
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
		ResponseBody:    getResponseBody(c.Response().Body()),
	}

	latency := time.Since(c.Context().Time()).String()
	document.Latency = latency

	es := config.GetElasticConnection()
	_, err := es.Index("api-log", esutil.NewJSONReader(&document))
	if err != nil {
		fmt.Println("error store log to elastic", err)
		return
	}

	// TODO
	// Check is response 201 to make sure successfully store log to elastic
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

func getResponseBody(body []byte) map[string]interface{} {
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
