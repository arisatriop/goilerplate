package repository

import (
	"context"
	"goilerplate/internal/domain/store"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type storeRepo struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) store.Repository {
	return &storeRepo{
		db: db,
	}
}

func (r *storeRepo) WithTx(ctx context.Context) store.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewStore(tx)
	}
	return r
}

func (r *storeRepo) CreateStore(ctx context.Context, s *store.Store) (*store.Store, error) {
	var model model.Store
	model.ID = s.ID
	model.UserID = s.UserID
	model.Name = s.Name
	model.Desc = s.Desc
	model.Address = s.Address
	model.Phone = s.Phone
	model.Email = s.Email
	model.WebURL = s.WebURL
	model.IsActive = s.IsActive

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	return r.toDomainEntity(&model), nil
}

func (r *storeRepo) GetStoreByUserID(ctx context.Context, userID uuid.UUID) (*store.Store, error) {
	var model model.Store

	if err := r.db.WithContext(ctx).
		Select("id", "user_id", "name", "desc", "address", "phone", "email", "web_url", "is_active").
		Where("user_id = ?", userID).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEntity(&model), nil
}

func (r *storeRepo) GetStoreByID(ctx context.Context, id uuid.UUID) (*store.Store, error) {
	var model model.Store

	if err := r.db.WithContext(ctx).
		Select("id", "user_id", "name", "desc", "address", "phone", "email", "web_url", "is_active").
		Where("id = ? AND is_active = ? AND deleted_at IS NULL", id, true).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomainEntity(&model), nil
}

func (r *storeRepo) toDomainEntity(m *model.Store) *store.Store {
	if m == nil {
		return nil
	}
	return &store.Store{
		ID:       m.ID,
		UserID:   m.UserID,
		Name:     m.Name,
		Desc:     m.Desc,
		Address:  m.Address,
		Phone:    m.Phone,
		Email:    m.Email,
		WebURL:   m.WebURL,
		IsActive: m.IsActive,
	}
}
