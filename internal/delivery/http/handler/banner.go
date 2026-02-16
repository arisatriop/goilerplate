package handler

import (
	bannerapp "goilerplate/internal/application/banner"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	"goilerplate/internal/domain/banner"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"
	"goilerplate/pkg/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Banner struct {
	Validator  *validator.Validate
	Usecase    banner.Usecase
	AppService bannerapp.ApplicationService
}

func NewBanner(validator *validator.Validate, usecase banner.Usecase, appService bannerapp.ApplicationService) *Banner {
	return &Banner{
		Validator:  validator,
		Usecase:    usecase,
		AppService: appService,
	}
}

// * COMMAND

func (h *Banner) Create(ctx *fiber.Ctx) error {
	// Get store ID from context (set by store middleware)
	storeIDStr := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	storeID, _ := uuid.Parse(storeIDStr)

	// Parse multipart form
	form, err := ctx.MultipartForm()
	if err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	// Get files from form
	files := form.File["banners"] // assuming form field name is "banners"
	if len(files) == 0 {
		return response.BadRequest(ctx, "No banner files provided", nil)
	}

	// Validate file count (basic validation, more detailed validation in service)
	if len(files) > 10 { // arbitrary limit for request size
		return response.BadRequest(ctx, "Too many files. Maximum 10 files allowed per request", nil)
	}

	// Call application service to handle file upload and banner creation
	err = h.AppService.CreateBannerWithFiles(ctx.UserContext(), storeID, files)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil)
}

func (h *Banner) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	parseID, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequest(ctx, "invalid id", nil)
	}

	err = h.Usecase.Delete(ctx.UserContext(), parseID, request.ToBannerFilterByStore(ctx, false))
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil)
}

func (h *Banner) UpdateToggleActive(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	parseID, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequest(ctx, "invalid id", nil)
	}

	storeIDstr, _ := ctx.Locals(string(constants.ContextKeyStoreID)).(string)
	parseStoreID, _ := uuid.Parse(storeIDstr)

	result, err := h.AppService.UpdateToggleActive(ctx.UserContext(), &banner.Banner{
		ID:      parseID,
		StoreID: parseStoreID,
	})
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, result)

}

func (h *Banner) Upload(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	files := form.File["files"]
	if len(files) == 0 {
		return response.BadRequest(ctx, "At least one file is required", nil)
	}

	uploadedBanners, err := h.Usecase.Upload(ctx.UserContext(), files)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	bannerResponses := presenter.ToBannerUploadListResponse(uploadedBanners)

	return response.Success(ctx, bannerResponses, response.WithMessage("Files uploaded successfully"))
}

// * END OF COMMAND

// * QUERY

func (h *Banner) GetListByStore(ctx *fiber.Ctx) error {
	filter := request.ToBannerFilterByStore(ctx, true)
	results, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	bannerResponses := presenter.ToBannerListResponse(results)
	paginatedResponse := pagination.NewPaginatedResponse(bannerResponses, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse)
}

func (h *Banner) GetListGuest(ctx *fiber.Ctx) error {
	filter := request.ToBannerFilterByStore(ctx, true)
	filter.IsActive = utils.Pointer(true)
	results, total, err := h.Usecase.GetList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	bannerResponses := presenter.ToBannerListResponse(results)
	paginatedResponse := pagination.NewPaginatedResponse(bannerResponses, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse)
}

// * END OF QUERY
