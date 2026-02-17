package repository

import (
	"context"
	"goilerplate/internal/domain/role"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"

	"gorm.io/gorm"
)

type roleRepo struct {
	db *gorm.DB
}

func NewRole(db *gorm.DB) role.Repository {
	return &roleRepo{db: db}
}

func (r *roleRepo) WithTx(ctx context.Context) role.Repository {
	tx := transaction.GetTxFromContext(ctx)
	if tx != nil {
		return NewRole(tx)
	}
	return r
}

func (r *roleRepo) GetRoleBySlug(ctx context.Context, slug string) (*role.Role, error) {
	var rl model.Role
	if err := r.db.WithContext(ctx).
		Select("id", "name", "slug", "description").
		Where("slug = ?", slug).
		First(&rl).Error; err != nil {
		return nil, err
	}
	return r.toDomainEntity(&rl), nil
}

func (r *roleRepo) toDomainEntity(m *model.Role) *role.Role {
	if m == nil {
		return nil
	}
	return &role.Role{
		ID:          m.ID,
		Name:        m.Name,
		Slug:        m.Slug,
		Description: m.Description,
	}
}
