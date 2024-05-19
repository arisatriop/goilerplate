package middleware

import (
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Detail struct {
	Timestamp       string
	RequestId       interface{}
	Ip              string
	Method          string
	BaseUrl         string
	Endpoint        string
	OriginalUrl     string
	Headers         string
	RequestContent  interface{}
	RequestBody     interface{}
	Status          int
	ResponseContent interface{}
	ResponseBody    string
	Latency         time.Duration
}

func Log(app *fiber.App) {

	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()

		pushToElastic(c)

		return err
	})
}

func pushToElastic(c *fiber.Ctx) {
	request := map[string]interface{}{
		"timestamp":       c.Context().Time().Format("2006-01-02 15:04:05.000"),
		"requestId":       c.Locals("requestid"),
		"ip":              c.IP(),
		"method":          c.Method(),
		"baseUrl":         c.BaseURL(),
		"endpoint":        c.Path(),
		"originalUrl":     c.BaseURL() + c.OriginalURL(),
		"headers":         c.Request().Header.String(),
		"requestContent":  c.Request(),
		"requestBody":     getRequestBody(c.Request()),
		"status":          c.Response().StatusCode(),
		"responseContent": c.Response(),
		"responseBody":    string(c.Response().Body()),
	}

	latency := time.Since(c.Context().Time())
	request["latency"] = latency

	// TODO
	// push to elastic
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
