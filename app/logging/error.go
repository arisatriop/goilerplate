package logging

import (
	"fmt"
	"goilerplate/config"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
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
	Store(c *fiber.Ctx, message string) error
	StoreToFile(doc *ErrorDocument) error
	GetDocument(c *fiber.Ctx) *ErrorDocument
}

type ErrorLogImpl struct{}

func NewErrorLog() ErrorLog {
	return &ErrorLogImpl{}
}

func (log *ErrorLogImpl) Store(c *fiber.Ctx, message string) error {

	document := log.GetDocument(c)
	document.Message = message

	log.StoreToFile(document)

	var err error
	var res *esapi.Response

	app := config.GetAppVariable()
	if app.LogChannel != "elasticsearch" {
		res, err = app.ElasticClient.Index("error-log", esutil.NewJSONReader(&document))
		if err != nil {
			return fmt.Errorf("error %v", err)
		}
		defer res.Body.Close()
	}

	return nil
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

func (log *ErrorLogImpl) StoreToFile(document *ErrorDocument) error {
	file, err := os.OpenFile("./logs/error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error open file: ", err.Error())
		return nil
	}

	data := fmt.Sprintf("%s | %s | %s | %s | %s\n%s\n",
		document.Timestamp.Format("2006-01-02 15:04:05.000"),
		document.RequestId, document.RequestHeaders["X-User"][0],
		document.Method, document.Endpoint, document.Message)

	if _, err := file.WriteString(data + "\n"); err != nil {
		fmt.Println("error store log to file: ", err.Error())
		return nil
	}

	return nil // always return nil
}
