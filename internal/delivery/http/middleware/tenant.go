package middleware

import (
	"context"
	"fmt"
	"goilerplate/config"
	"goilerplate/internal/domain/store"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/crypto"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Store struct {
	cfg       *config.Config
	storeRepo store.Repository
}

func NewStore(cfg *config.Config, storeRepo store.Repository) *Store {
	return &Store{
		cfg:       cfg,
		storeRepo: storeRepo,
	}
}

func (m *Store) GetStore() fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		userID, err := uuid.Parse(ctx.Locals(string(constants.ContextKeyUserID)).(string))
		if err != nil {
			return response.HandleError(ctx, err)
		}

		store, err := m.storeRepo.GetStoreByUserID(ctx.UserContext(), userID)
		if err != nil {
			return response.HandleError(ctx, err)
		}

		storeIDCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyStoreID, store.ID.String())
		ctx.SetUserContext(storeIDCtx)
		ctx.Locals(string(constants.ContextKeyStoreID), store.ID.String())

		return ctx.Next()
	}
}

func (m *Store) SetTenantID() fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		tenantID := ctx.Get("X-Tenant-ID")
		if tenantID == "" {
			return response.BadRequest(ctx, "Invalid tenant id", nil)
		}

		decryptTenantID, err := crypto.DecryptString(tenantID, m.cfg.Crypto.EncryptionKey)
		if err != nil {
			fmt.Println("error: ", err)
			return response.HandleError(ctx, err)
		}

		tenantIDCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyStoreID, decryptTenantID)
		ctx.SetUserContext(tenantIDCtx)
		ctx.Locals(string(constants.ContextKeyStoreID), decryptTenantID)

		return ctx.Next()
	}
}
