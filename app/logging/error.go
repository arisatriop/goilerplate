package logging

import (
	"fmt"
	"goilerplate/config"
	"os"
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

type ErrorLog interface {
	Store(c *fiber.Ctx, message string)
	StoreToFiles(doc *ErrorDocument)
	GetDocument(c *fiber.Ctx) *ErrorDocument
}

type ErrorLogImpl struct{}

func NewErrorLog() ErrorLog {
	return &ErrorLogImpl{}
}

func (log *ErrorLogImpl) Store(c *fiber.Ctx, message string) {

	document := log.GetDocument(c)
	document.Message = message

	fmt.Println("header: ", document.RequestHeaders)

	log.StoreToFiles(document)

	es := config.GetElasticConnection()
	res, err := es.Index("error-log", esutil.NewJSONReader(&document))
	if err != nil {
		fmt.Println("Error store Error Log to elastic: ", err)
	}
	defer res.Body.Close()

}

func (log *ErrorLogImpl) GetDocument(c *fiber.Ctx) *ErrorDocument {

	api := NewApiLog()

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

func (log *ErrorLogImpl) StoreToFiles(document *ErrorDocument) {
	file, _ := os.OpenFile("./logs/error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	data := fmt.Sprintf("%s | %s | %s | %s | %s\n%s\n",
		document.Timestamp.Format("2006-01-02 15:04:05.000"), document.RequestId, document.RequestHeaders["X-User"][0], document.Method, document.Endpoint, document.Message)

	file.WriteString(data + "\n")
}
