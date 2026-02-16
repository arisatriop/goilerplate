package handler

import (
	"goilerplate/internal/application/productapp"
	dtorequest "goilerplate/internal/delivery/http/dto/request"
	"goilerplate/internal/delivery/http/presenter"
	"goilerplate/internal/delivery/http/request"
	"goilerplate/internal/domain/product"
	"goilerplate/internal/domain/productimage"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/pagination"
	"goilerplate/pkg/response"
	"goilerplate/pkg/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Product struct {
	Validator  *validator.Validate
	AppService productapp.ApplicationService
	Usecase    product.Usecase
}

func NewProduct(validator *validator.Validate, appService productapp.ApplicationService, usecase product.Usecase) *Product {
	return &Product{
		Validator:  validator,
		AppService: appService,
		Usecase:    usecase,
	}
}

// * COMMAND

func (h *Product) Create(ctx *fiber.Ctx) error {
	var req dtorequest.CreateProductsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	products := make([]productapp.Product, len(req.Products))
	for i, p := range req.Products {
		price, err := decimal.NewFromString(p.Price)
		if err != nil {
			return response.BadRequest(ctx, "Invalid price format", nil)
		}
		products[i] = productapp.Product{
			Name:       p.Name,
			Desc:       p.Desc,
			Price:      price,
			Categories: p.CategoryIDs,
		}
		productImages := make([]productimage.ProductImage, len(p.Images))
		for j, img := range p.Images {
			productImages[j] = productimage.ProductImage{
				FileType:    img.FileType,
				FileStorage: img.FileStorage,
				FileName:    img.FileName,
				FilePath:    img.FilePath,
				FileURL:     img.FileURL,
				IsPrimary:   img.IsPrimary,
			}
		}
		products[i].Images = productImages
	}

	err := h.AppService.CreateProducts(ctx.UserContext(), products)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil)
}

func (h *Product) CreateProductImages(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	var req dtorequest.ProductImagesRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	images := make([]productimage.ProductImage, len(req.Images))
	for i, img := range req.Images {
		images[i] = productimage.ProductImage{
			FileType:    img.FileType,
			FileStorage: img.FileStorage,
			FileName:    img.FileName,
			FilePath:    img.FilePath,
			FileURL:     img.FileURL,
			IsPrimary:   img.IsPrimary,
		}
	}

	err = h.AppService.CreateProductImages(ctx.UserContext(), id, images)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil)
}

func (h *Product) CreateProductCategory(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	var req dtorequest.CreateProductCategoriesRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	err = h.AppService.CreateProductCategories(ctx.UserContext(), id, req.CategoryIDs)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Created(ctx, nil)
}

func (h *Product) UpdateProductCategory(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	var req dtorequest.CreateProductCategoriesRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	err = h.AppService.UpdateProductCategories(ctx.UserContext(), id, req.CategoryIDs)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil)
}

func (h *Product) Update(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	var req dtorequest.UpdateProductRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	if err := h.Validator.Struct(&req); err != nil {
		validationErrors := response.FormatValidationErrors(err)
		return response.ValidationError(ctx, validationErrors)
	}

	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		return response.BadRequest(ctx, "Invalid price format", nil)
	}

	prod := &product.Product{
		ID:          id,
		Name:        req.Name,
		Description: req.Desc,
		Price:       price,
	}

	if err := h.Usecase.UpdateProduct(ctx.UserContext(), prod); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, nil)
}

func (h *Product) UpdateToggleActive(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	result, err := h.AppService.UpdateToggleActive(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, result)
}

func (h *Product) UpdateToggleAvailable(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	result, err := h.Usecase.UpdateToggleAvailable(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, result)
}

func (h *Product) Delete(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	err = h.AppService.DeleteProductByID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

func (h *Product) DeleteProductImageByID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	err = h.AppService.DeleteProductImageByID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

func (h *Product) DeleteProductCategoryByID(ctx *fiber.Ctx) error {
	pidStr := ctx.Params("id")
	pid, err := uuid.Parse(pidStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	cidStr := ctx.Params("cid")
	cid, err := uuid.Parse(cidStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid category ID", nil)
	}

	err = h.AppService.DeleteProductCategoryByID(ctx.UserContext(), pid, cid)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

func (h *Product) DeleteProductImagesByProductID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	err = h.AppService.DeleteProductImagesByProductID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

func (h *Product) DeleteProductCategoryByProductID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	err = h.AppService.DeleteProductCategoryByProductID(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.NoContent(ctx)
}

func (h *Product) MarkImageAsPrimary(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	result, err := h.AppService.MarkImagesAsPrimary(ctx.UserContext(), id)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, result)
}

// * END OF COMMAND

// * QUERY

func (h *Product) GetDetails(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(ctx, "Invalid product ID", nil)
	}

	filter := request.ToProductFilter(ctx, false)
	filter.ProductID = &id

	products, err := h.AppService.GetProductDetails(ctx.UserContext(), id, *filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	resp := presenter.ToProductDetailsResponse(products)

	return response.Success(ctx, resp)

}

func (h *Product) GetMenuProductList(ctx *fiber.Ctx) error {
	filter := request.ToProductFilter(ctx, true)
	filter.IsActive = utils.Pointer(true)

	products, total, err := h.Usecase.GetMenuProductList(ctx.UserContext(), filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	result := presenter.ToProductListResponse(products)
	paginatedResponse := pagination.NewPaginatedResponse(result, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse)
}

func (h *Product) GetListWithCategories(ctx *fiber.Ctx) error {
	filter := request.ToProductFilter(ctx, true)

	products, total, err := h.AppService.GetProductListWithCategory(ctx.UserContext(), *filter)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	productResponses := presenter.ToProductListWithCategoriesResponse(products)
	paginatedResponse := pagination.NewPaginatedResponse(productResponses, total, filter.Pagination.Page, filter.Pagination.Limit)

	return response.Success(ctx, paginatedResponse)

}

func (h *Product) GetCategoryListWithProductsByGuest(ctx *fiber.Ctx) error {
	data, err := h.AppService.GetCategoryListWithProducts(ctx.UserContext(), *request.ToCategoryWithProductsFilter(ctx))
	if err != nil {
		return response.HandleError(ctx, err)
	}

	categoryResponses := presenter.ToCategoryWithProductsResponse(data)

	return response.Success(ctx, categoryResponses)
}

// * END OF QUERY
