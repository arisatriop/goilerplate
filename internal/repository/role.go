package repository

import (
	"context"
	"goilerplate/internal/entity"
	"goilerplate/internal/model/role"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Count(ctx context.Context, db *gorm.DB, req *role.GetRequest) (int64, error)
	Create(ctx context.Context, db *gorm.DB, role *entity.Role) error
	Update(ctx context.Context, db *gorm.DB, role *entity.Role) error
	GetAll(ctx context.Context, db *gorm.DB, req *role.GetRequest) ([]entity.Role, int64, error)
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, db *gorm.DB, name string) (*entity.Role, error)
	GetByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) ([]entity.Role, error)
	BatchInsert(ctx context.Context, db *gorm.DB, role []entity.Role) error
}

type roleRepository struct {
	Log *logrus.Logger
}

func NewRoleRepository(log *logrus.Logger) RoleRepository {
	return &roleRepository{
		Log: log,
	}
}

func (r *roleRepository) Create(ctx context.Context, db *gorm.DB, role *entity.Role) error {
	if err := db.Create(role).Error; err != nil {
		r.Log.Error("failed to create role: ", err)
		return err
	}
	return nil
}

func (r *roleRepository) Update(ctx context.Context, db *gorm.DB, role *entity.Role) error {
	if err := db.Save(role).Error; err != nil {
		r.Log.Error("failed to update role: ", err)
		return err
	}
	return nil
}

func (r *roleRepository) GetAll(ctx context.Context, db *gorm.DB, req *role.GetRequest) ([]entity.Role, int64, error) {
	query := db.Model(&entity.Role{}).Where("deleted_at IS NULL")

	if req.Keyword != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(req.Keyword)+"%")
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	var roles []entity.Role
	if err := query.Find(&roles).Error; err != nil {
		r.Log.Error("failed to get roles: ", err)
		return nil, 0, err
	}

	count, err := r.Count(ctx, db, req)
	if err != nil {
		return nil, 0, err
	}

	return roles, count, nil
}

func (r *roleRepository) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	if err := db.Where("id = ? and deleted_at is null", id).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No role found with the given ID
		}
		r.Log.Error("failed to get role by ID: ", err)
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, db *gorm.DB, name string) (*entity.Role, error) {
	var role entity.Role
	if err := db.Where("name = ? and deleted_at is null", name).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No role found with the given name
		}
		r.Log.Error("failed to get role by name: ", err)
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) ([]entity.Role, error) {
	var roles []entity.Role
	if err := db.Table("roles").
		Joins("JOIN role_users ON roles.id = role_users.role_id").
		Where("role_users.user_id = ? AND roles.deleted_at IS NULL", userID).
		Find(&roles).Error; err != nil {
		r.Log.Error("failed to get roles by user ID: ", err)
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) BatchInsert(ctx context.Context, db *gorm.DB, role []entity.Role) error {
	if err := db.CreateInBatches(&role, 100).Error; err != nil {
		r.Log.Error("failed to create roles: ", err)
		return err
	}
	return nil
}

func (r *roleRepository) Count(ctx context.Context, db *gorm.DB, req *role.GetRequest) (int64, error) {
	query := db.Model(&entity.Role{}).Where("deleted_at IS NULL")

	if req.Keyword != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(req.Keyword)+"%")
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.Log.Error("failed to count roles: ", err)
		return 0, err
	}

	return count, nil
}
