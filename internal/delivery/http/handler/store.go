package handler

import (
	dtoresponse "goilerplate/internal/delivery/http/dto/response"
	"goilerplate/internal/domain/store"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Store struct {
	usecase store.Usecase
}

func NewStore(usecase store.Usecase) *Store {
	return &Store{
		usecase: usecase,
	}
}

func (h *Store) GetStoreInfo(ctx *fiber.Ctx) error {
	storeID := ctx.Locals("store_id").(string)

	store, err := h.usecase.GetInfo(ctx.UserContext(), storeID)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	result := &dtoresponse.StoreResponse{
		ID:          storeID,
		Name:        store.Name,
		Description: store.Desc,
		Address:     store.Address,
		Phone:       store.Phone,
		Email:       store.Email,
		WebURL:      store.WebURL,
	}

	return response.Success(ctx, result)
}
