package usecase

import (
	"context"
	"goilerplate/config"
	"goilerplate/internal/entity"
	"goilerplate/internal/model/menu"
	"goilerplate/internal/repository"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type MenuUsecase interface {
	GetAll(ctx context.Context, req *menu.GetRequest) ([]menu.GetAllResponse, int64, error)
	BuildMenuTree(ctx context.Context, menus []entity.Menu, parentID *uuid.UUID) ([]menu.GetAllResponse, error)
}

type menuUsecase struct {
	Log            *logrus.Logger
	DB             *config.DB
	MenuRepository repository.MenuRepository
}

func NewMenuUsecase(log *logrus.Logger, db *config.DB, menuRepo repository.MenuRepository) MenuUsecase {
	return &menuUsecase{
		Log:            log,
		DB:             db,
		MenuRepository: menuRepo,
	}
}

func (u *menuUsecase) GetAll(ctx context.Context, req *menu.GetRequest) ([]menu.GetAllResponse, int64, error) {
	allMenus, err := u.MenuRepository.GetAll(ctx, u.DB.GDB.WithContext(ctx), req)
	if err != nil {
		return nil, 0, err
	}

	menuThree, err := u.BuildMenuTree(ctx, allMenus, nil)
	if err != nil {
		return nil, 0, err
	}

	filteredMenuThree := func(menuThree []menu.GetAllResponse) []menu.GetAllResponse {
		var menus []menu.GetAllResponse
		for _, v := range menuThree {
			if req.Keyword == "" || strings.Contains(strings.ToLower(v.Name), strings.ToLower(req.Keyword)) {
				menus = append(menus, v)
			}
		}
		return menus
	}(menuThree)

	return filteredMenuThree, int64(len(filteredMenuThree)), nil
}

func (u *menuUsecase) BuildMenuTree(ctx context.Context, menus []entity.Menu, parentID *uuid.UUID) ([]menu.GetAllResponse, error) {
	var tree []menu.GetAllResponse

	for _, mn := range menus {
		if (mn.ParentID == nil && parentID == nil) || (mn.ParentID != nil && parentID != nil && *mn.ParentID == *parentID) {
			children, err := u.BuildMenuTree(ctx, menus, &mn.ID)
			if err != nil {
				return nil, err
			}

			if len(children) == 0 {
				children = []menu.GetAllResponse{}
			}

			node := menu.GetAllResponse{
				ID:       mn.ID,
				Name:     mn.Name,
				Path:     mn.Path,
				Icon:     "",
				Order:    0,
				IsActive: mn.IsActive,
				Child:    children,
			}
			if mn.Icon != nil {
				node.Icon = *mn.Icon
			}
			if mn.Order != nil {
				node.Order = *mn.Order
			}

			permission := []menu.Permission{}
			menuPermission, err := u.MenuRepository.GetPermission(ctx, u.DB.GDB.WithContext(ctx), node.ID)
			if err != nil {
				return nil, err
			}
			for _, v := range menuPermission {
				permission = append(permission, menu.Permission{
					ID:    v.ID.String(),
					Name:  v.Permission,
					Order: v.Order,
				})
			}

			node.Permission = permission
			tree = append(tree, node)
		}
	}

	// ✅ Sort tree by Order ascending
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Order < tree[j].Order
	})

	return tree, nil
}
