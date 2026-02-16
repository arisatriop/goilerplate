package handler

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/productimage"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ProductImage struct {
	Validator *validator.Validate
	Usecase   productimage.Usecase
}

func NewProductImage(validator *validator.Validate, usecase productimage.Usecase) *ProductImage {
	return &ProductImage{
		Validator: validator,
		Usecase:   usecase,
	}
}

func (h *ProductImage) Upload(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return response.BadRequest(ctx, constants.MsgInvalidRequestBody, nil)
	}

	files := form.File["files"]
	if len(files) == 0 {
		return response.BadRequest(ctx, "At least one file is required", nil)
	}

	uploadedProductImages, err := h.Usecase.Upload(ctx.UserContext(), files)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	var productImageResponses []dtoresponse.ProductImageUploadResponse
	for _, b := range uploadedProductImages {
		productImageResponses = append(productImageResponses, dtoresponse.ProductImageUploadResponse{
			Filetype:    b.FileType,
			FileStorage: b.FileStorage,
			Filename:    b.FileName,
			Filepath:    b.FilePath,
			FileURL:     b.FileURL,
		})
	}

	return response.Success(ctx, productImageResponses, response.WithMessage("Files uploaded successfully"))
}
