package handler

import (
	"goilerplate/internal/application/category"
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	domaincategory "goilerplate/internal/domain/category"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"
	"goilerplate/pkg/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Category struct {
	Validator  *validator.Validate
	AppService category.ApplicationService
	Usecase    domaincategory.Usecase
}

func NewCategory(validator *validator.Validate, appService category.ApplicationService, usecase domaincategory.Usecase) *Category {
	return &Category{
		Validator:  validator,
		AppService: appService,
		Usecase:    usecase,
	}
}

// * COMMAND

func (h *Category) Create(ctx *fiber.Ctx) error {
	var req dtorequest.CategoryCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	// Get store ID from context (set by store middleware)
	storeIDStr := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	storeID, _ := uuid.Parse(storeIDStr)

	// Convert request names to category entities
	categories := make([]*domaincategory.Category, len(req.Name))
	for i, name := range req.Name {
		categories[i] = &domaincategory.Category{
			Name:    name,
			StoreID: storeID,
		}
	}

	// Call application service to create categories
	err := h.AppService.CreateCategory(ctx.UserContext(), categories)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil)
}

func (h *Category) UpdateToggleActive(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	parseID, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequest(ctx, "invalid id", nil)
	}

	storeIDstr, _ := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	parseStoreID, _ := uuid.Parse(storeIDstr)

	result, err := h.AppService.UpdateToggleActive(ctx.UserContext(), &domaincategory.Category{
		ID:      parseID,
		StoreID: parseStoreID,
	})
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, result)
}

func (h *Category) SoftDelete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequest(ctx, "invalid id", nil)
	}

	err = h.Usecase.SoftDelete(ctx.UserContext(), parsedID, request.ToCategoryFilterByStore(ctx, false))
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil)
}

// * END OF COMMAND

// * QUERY

func (h *Category) GetListByGuest(ctx *fiber.Ctx) error {
	var req dtorequest.CategoryListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return utils.Error(http.StatusBadRequest, constants.MsgInvalidRequestBody)
	}

	filter := request.ToCategoryFilterByStore(ctx, true)
	filter.IsActive = utils.Pointer(true)
	result, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	categoryResponse := presenter.ToCategoryListResponse(result)
	paginatedResponse := pagination.NewPaginatedResponse(categoryResponse, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse)
}

func (h *Category) GetListByStore(ctx *fiber.Ctx) error {
	filter := request.ToCategoryFilterByStore(ctx, true)
	result, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	categoryResponse := presenter.ToCategoryListResponse(result)
	paginatedResponse := pagination.NewPaginatedResponse(categoryResponse, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse)
}

func (h *Category) GetMenuCategoryList(ctx *fiber.Ctx) error {
	filter := request.ToCategoryFilterByStore(ctx, false)
	filter.IsActive = utils.Pointer(true)

	result, _, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	categoryResponse := presenter.ToCategoryListResponse(result)

	return response.Success(ctx, categoryResponse)
}

// * END OF QUERY
