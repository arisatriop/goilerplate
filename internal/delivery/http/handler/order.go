package handler

import (
	"goilerplate/internal/application/orderapp"
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	"goilerplate/internal/domain/order"
	"goilerplate/internal/domain/orderitem"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

type Order struct {
	Validator  *validator.Validate
	AppService orderapp.ApplicationService
	Usecase    order.Usecase
}

func NewOrder(validator *validator.Validate, appService orderapp.ApplicationService, orderUC order.Usecase) *Order {
	return &Order{
		Validator:  validator,
		AppService: appService,
		Usecase:    orderUC,
	}
}

// * Query
func (h *Order) GetList(ctx *fiber.Ctx) error {
	filter := request.ToOrderFilter(ctx)

	result, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	orderListResponse := presenter.ToOrderSummaryListResponse(result)
	paginatedResponse := pagination.NewPaginatedResponse(orderListResponse, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse)
}

func (h *Order) GetDetail(ctx *fiber.Ctx) error {
	orderID := ctx.Params("id")
	if orderID == "" {
		return response.NotFound(ctx, "")
	}

	storeID := ctx.Locals(string(constants.ContextKeyStoreID)).(string)

	result, err := h.AppService.GetOrderDetail(ctx.UserContext(), orderID, storeID)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, presenter.ToOrderDetailResponse(result))
}

// * Command
func (h *Order) CreateOrder(ctx *fiber.Ctx) error {
	var req dtorequest.OrderCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	amount, err := decimal.NewFromString(req.GrandTotal)
	if err != nil {
		return response.BadRequest(ctx, "invalid amount format", nil)
	}

	order := order.Order{
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CutomerEmail,
		CustomerPhone: req.CutomerPhone,
		TableNumber:   req.TableNumber,
		Notes:         req.Notes,
		Amount:        amount,
		StoreID:       ctx.Locals(string(constants.ContextKeyStoreID)).(string),
	}

	orderItems := make([]orderitem.OrderItem, len(req.Items))
	for i, item := range req.Items {
		price, err := decimal.NewFromString(item.Price)
		if err != nil {
			return response.BadRequest(ctx, "invalid price format", nil)
		}
		subTotal, err := decimal.NewFromString(item.SubTotal)
		if err != nil {
			return response.BadRequest(ctx, "invalid sub total format", nil)
		}
		orderItems[i] = orderitem.OrderItem{
			ProductID: item.ProductID,
			Name:      item.ProductName,
			Price:     price,
			Image:     item.Image,
			Qty:       item.Quantity,
			SubTotal:  subTotal,
			Note:      item.Notes,
		}
	}

	createdOrder, err := h.AppService.CreateOrder(ctx.Context(), orderapp.Order{
		Order:      &order,
		OrderItems: orderItems,
	})
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, presenter.ToOrderCreateResponse(createdOrder, orderItems))

}
