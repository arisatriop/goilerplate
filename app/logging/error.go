package logging

import (
	"fmt"
	"goilerplate/config"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/gofiber/fiber/v2"
)

type ErrorDocument struct {
	Timestamp      time.Time           `json:"timestamp"`
	RequestId      interface{}         `json:"request_id"`
	Method         string              `json:"method"`
	Endpoint       string              `json:"endpoint"`
	OriginalUrl    string              `json:"original_url"`
	Message        string              `json:"message"`
	RequestHeaders map[string][]string `json:"request_headers"`
	RequestBody    interface{}         `json:"request_body"`
}

type ErrorLog struct{}

func (log *ErrorLog) Store(c *fiber.Ctx, message string) {

	document := log.GetDocument(c)
	document.Message = message

	es := config.GetElasticConnection()
	res, err := es.Index("error-log", esutil.NewJSONReader(&document))
	if err != nil {
		fmt.Println("Error store Error Log to elastic: ", err)
	}
	defer res.Body.Close()

}

func (log *ErrorLog) GetDocument(c *fiber.Ctx) *ErrorDocument {

	api := &ApiLog{}

	return &ErrorDocument{
		Timestamp:      c.Context().Time(),
		RequestId:      c.Locals("requestid"),
		Method:         c.Method(),
		Endpoint:       c.Path(),
		OriginalUrl:    c.BaseURL() + c.OriginalURL(),
		RequestHeaders: c.GetReqHeaders(),
		RequestBody:    api.GetRequestBody(c.Request()),
	}
}
