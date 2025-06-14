package usecase

import (
	"context"
	"goilerplate/internal/config"
	"goilerplate/internal/model/role"
	"goilerplate/internal/repository"
	"goilerplate/pkg/helper"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type RoleUseCase interface {
	Create(ctx context.Context, req *role.CreateRequest) error
	Update(ctx context.Context, id uuid.UUID, req *role.UpdateRequest) error
	Delete(ctx context.Context, req *role.DeleteRequest) error
	GetAll(ctx context.Context, req *role.GetRequest) ([]role.GetAllResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*role.GetResponse, error)
}

type roleUsecase struct {
	Log              *logrus.Logger
	DB               *config.DB
	RoleRepo         repository.RoleRepository
	MenuPermRepo     repository.MenuPermissionRepository
	MenuPermRoleRepo repository.MenuPermissionRoleRepository
}

func NewRoleUseCase(
	log *logrus.Logger,
	db *config.DB,
	roleRepo repository.RoleRepository,
	menuPermRepo repository.MenuPermissionRepository,
	menuPermRoleRepo repository.MenuPermissionRoleRepository) RoleUseCase {
	return &roleUsecase{
		Log:              log,
		DB:               db,
		RoleRepo:         roleRepo,
		MenuPermRepo:     menuPermRepo,
		MenuPermRoleRepo: menuPermRoleRepo,
	}
}

func (u *roleUsecase) Create(ctx context.Context, req *role.CreateRequest) error {

	roleEntity := req.ToCreate()
	if err := u.ValidateCreateInput(ctx, req); err != nil {
		return err
	}

	tx := u.DB.GDB.WithContext(ctx).Begin()

	err := u.RoleRepo.Create(ctx, tx, roleEntity)
	if err != nil {
		tx.Rollback()
		return err
	}

	menuPermRoleEntity := req.ToCreatePermission(roleEntity.ID)
	if len(menuPermRoleEntity) > 0 {
		if err := u.MenuPermRoleRepo.BatchInsert(ctx, tx, menuPermRoleEntity); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Errorf("failed to commit transaction: %v\n", err)
		tx.Rollback()
		return err
	}

	return nil
}

func (u *roleUsecase) Update(ctx context.Context, id uuid.UUID, req *role.UpdateRequest) error {

	role, err := u.RoleRepo.GetByID(ctx, u.DB.GDB.WithContext(ctx), id)
	if err != nil {
		return err
	}

	if role == nil {
		return helper.Error(http.StatusNotFound, "Role not found")
	}

	req.ToUpdate(role)
	if err := u.ValidateUpdateInput(ctx, req); err != nil {
		return err
	}

	tx := u.DB.GDB.WithContext(ctx).Begin()
	if err := u.RoleRepo.Update(ctx, tx, role); err != nil {
		tx.Rollback()
		return err
	}

	err = u.MenuPermRoleRepo.HardDeleteByRoleID(ctx, tx, role.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	menuPermRoleEntity := req.ToCreatePermission(role.ID)
	if len(menuPermRoleEntity) > 0 {
		if err := u.MenuPermRoleRepo.BatchInsert(ctx, tx, menuPermRoleEntity); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Errorf("failed to commit transaction: %v\n", err)
		tx.Rollback()
		return err
	}

	return nil
}

func (u *roleUsecase) Delete(ctx context.Context, req *role.DeleteRequest) error {
	role, err := u.RoleRepo.GetByID(ctx, u.DB.GDB.WithContext(ctx), req.ID)
	if err != nil {
		return err
	}
	if role == nil {
		return helper.Error(http.StatusNotFound, "Role not found")
	}

	tx := u.DB.GDB.WithContext(ctx).Begin()

	req.ToDelete(role)
	if err := u.RoleRepo.Update(ctx, tx, role); err != nil {
		tx.Rollback()
		return err
	}

	if err := u.MenuPermRoleRepo.DeleteByRoleID(ctx, tx, role.ID, req.DeletedBy); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Errorf("failed to commit transaction: %v\n", err)
		tx.Rollback()
		return err
	}

	return nil
}

func (u *roleUsecase) GetAll(ctx context.Context, req *role.GetRequest) ([]role.GetAllResponse, int64, error) {
	roles, total, err := u.RoleRepo.GetAll(ctx, u.DB.GDB.WithContext(ctx), req)
	if err != nil {
		return nil, 0, err
	}

	var response []role.GetAllResponse
	for _, r := range roles {
		response = append(response, *role.ToGetAll(&r))
	}

	return response, total, nil
}

func (u *roleUsecase) GetByID(ctx context.Context, id uuid.UUID) (*role.GetResponse, error) {
	roles, err := u.RoleRepo.GetByID(ctx, u.DB.GDB.WithContext(ctx), id)
	if err != nil {
		return nil, err
	}

	if roles == nil {
		return nil, helper.Error(http.StatusNotFound, "Role not found")
	}

	return role.ToGet(roles), nil
}

func (u *roleUsecase) ValidateCreateInput(ctx context.Context, req *role.CreateRequest) error {

	role, err := u.RoleRepo.GetByName(ctx, u.DB.GDB.WithContext(ctx), req.Name)
	if err != nil {
		return err
	}

	if role != nil {
		return helper.Error(http.StatusBadRequest, "Role name already exists")
	}

	if err := u.ValidateInputPermission(ctx, req.Permission); err != nil {
		return err
	}

	return nil
}

func (u *roleUsecase) ValidateUpdateInput(ctx context.Context, req *role.UpdateRequest) error {
	role, err := u.RoleRepo.GetByName(ctx, u.DB.GDB.WithContext(ctx), req.Name)
	if err != nil {
		return err
	}

	if role != nil {
		if role.Name == req.Name && role.ID != req.ID {
			return helper.Error(http.StatusBadRequest, "Role name already exists")
		}
	}

	if err := u.ValidateInputPermission(ctx, req.Permission); err != nil {
		return err
	}

	return nil

}

func (u *roleUsecase) ValidateInputPermission(ctx context.Context, permission []string) error {
	// ! Validate menu permission id input
	for _, id := range permission {
		_ = id
	}
	return nil
}
