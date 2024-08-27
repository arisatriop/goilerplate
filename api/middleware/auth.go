package middleware

import (
	"context"
	"goilerplate/config"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func Auth(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {

		authorization := strings.Replace(c.Get("Authorization"), "Bearer ", "", -1)

		_, err := getToken("token_" + authorization)
		if err == redis.Nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal Server Error",
			})
		}

		return c.Next()
	})
}

func getToken(token string) (string, error) {
	redisClient := config.GetAppVariable().RedisClient
	return redisClient.Get(context.Background(), token).Result()

}
