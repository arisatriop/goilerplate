package middleware

import (
	"fmt"
	"goilerplate/app/logging"

	"github.com/gofiber/fiber/v2"
)

func Log(app *fiber.App) {

	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()

		log := logging.NewApiLog()
		errr := log.Store(c)
		if errr != nil {
			fmt.Println("error store API Log: ", err)
			errorLog := logging.NewErrorLog()
			errorLog.Store(c, errr.Error())
		}

		return err
	})
}

// func pushToElastic(c *fiber.Ctx) {
// 	// document := Document{
// 	// 	Timestamp:          c.Context().Time(),
// 	// 	RequestId:          c.Locals("requestid"),
// 	// 	Ip:                 c.IP(),
// 	// 	Method:             c.Method(),
// 	// 	BaseUrl:            c.BaseURL(),
// 	// 	Endpoint:           c.Path(),
// 	// 	OriginalUrl:        c.BaseURL() + c.OriginalURL(),
// 	// 	RequestHeaders:     c.GetReqHeaders(),
// 	// 	RequestBody:        getRequestBody(c.Request()),
// 	// 	Status:             c.Response().StatusCode(),
// 	// 	ResponseHeaders:    c.GetRespHeaders(),
// 	// 	ResponseBody:       getResponseBody(c.Response().Body()),
// 	// 	ResponseBodyString: string(c.Response().Body()),
// 	// }

// 	// latency := time.Since(c.Context().Time()).String()
// 	// document.Latency = latency

// 	// es := config.GetElasticConnection()
// 	// _, err := es.Index("api-log", esutil.NewJSONReader(&document))
// 	// if err != nil {
// 	// 	fmt.Println("error store log to elastic", err)
// 	// 	return
// 	// }

// 	// TODO
// 	// Check is response 201 to make sure successfully store log to elastic
// }
