package usecase

import (
	"context"
	"fmt"
	"goilerplate/config"
	"goilerplate/internal/model/menupermission"
	"goilerplate/internal/repository"

	"github.com/sirupsen/logrus"
)

type MenuPermissionUsecase interface {
	GetAll(ctx context.Context, req *menupermission.GetRequest) ([]menupermission.GetAllResponse, int64, error)
}

type menuPermissionUsecase struct {
	Log                *logrus.Logger
	DB                 *config.DB
	MenuPermissionRepo repository.MenuPermissionRepository
}

func NewMenuPermissionUsecase(log *logrus.Logger, db *config.DB, menuPermissionRepo repository.MenuPermissionRepository) MenuPermissionUsecase {
	return &menuPermissionUsecase{
		Log:                log,
		DB:                 db,
		MenuPermissionRepo: menuPermissionRepo,
	}
}

func (uc *menuPermissionUsecase) GetAll(ctx context.Context, req *menupermission.GetRequest) ([]menupermission.GetAllResponse, int64, error) {

	result, total, err := uc.MenuPermissionRepo.GetAll(ctx, uc.DB.GDB.WithContext(ctx), req)
	if err != nil {
		return nil, 0, err
	}

	var response []menupermission.GetAllResponse
	for _, menuperm := range result {
		fmt.Println("Menu Permission:", menuperm.Menu)
		response = append(response, *menupermission.ToGetAll(&menuperm))
	}

	return response, total, nil
}
