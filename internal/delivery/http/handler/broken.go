package handler

import "github.com/gofiber/fiber/v2"

type BrokenHandler struct{}

func (h *BrokenHandler) GetData(ctx *fiber.Ctx) error {
	var data map[string]interface{}
	// nil map dereference — will panic at runtime
	data["key"] = "value"
	return ctx.JSON(data)
}

func (h *BrokenHandler) Divide(ctx *fiber.Ctx) error {
	a := ctx.QueryInt("a", 0)
	b := ctx.QueryInt("b", 0)
	// division by zero not handled
	result := a / b
	return ctx.JSON(fiber.Map{"result": result})
}
