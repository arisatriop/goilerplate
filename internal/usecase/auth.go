package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"goilerplate/internal/config"
	"goilerplate/internal/entity"
	"goilerplate/internal/model"
	"goilerplate/internal/model/auth"
	"goilerplate/internal/model/menu"
	"goilerplate/internal/repository"
	"goilerplate/pkg/helper"
	"net/http"
	"sort"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	Me(ctx context.Context, user *auth.User) (*auth.GetResponse, error)
	Token(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error)
	Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error)
	Logout(ctx context.Context, req *auth.LogoutRequest) error
	SetPermission(ctx context.Context, id uuid.UUID) error
	GetPermissions(ctx context.Context, id uuid.UUID) (model.Permission, error)
	GetPermissionFromRedis(ctx context.Context, key string) (map[string]struct{}, error)
}

type authUsecase struct {
	Config               *viper.Viper
	Log                  *logrus.Logger
	DB                   *config.DB
	MenuUsecase          MenuUsecase
	UserRepository       repository.UserRepository
	RoleRepo             repository.RoleRepository
	MenuRepo             repository.MenuRepository
	PermissionRepository repository.PermissionRepository
	RedisRepository      *repository.RedisRepository
}

func NewAuthUsecase(
	viper *viper.Viper,
	log *logrus.Logger,
	db *config.DB,
	menuUsecase MenuUsecase,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	menuRepo repository.MenuRepository,
	permissionRepo repository.PermissionRepository,
	redisRepo *repository.RedisRepository) AuthUsecase {
	return &authUsecase{
		Log:                  log,
		Config:               viper,
		DB:                   db,
		MenuUsecase:          menuUsecase,
		UserRepository:       userRepo,
		RoleRepo:             roleRepo,
		MenuRepo:             menuRepo,
		PermissionRepository: permissionRepo,
		RedisRepository:      redisRepo,
	}
}

func (u *authUsecase) Me(ctx context.Context, user *auth.User) (*auth.GetResponse, error) {
	db := u.DB.GDB.WithContext(ctx)

	role, err := u.RoleRepo.GetByUserID(ctx, db, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by user ID: %w", err)
	}

	allMenu, err := u.MenuRepo.GetAll(ctx, db, &menu.GetRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all menus: %w", err)
	}

	var myRoles []auth.Role
	for _, v := range role {
		menu, err := u.MenuRepo.GetByRoleID(ctx, db, v.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get menus by role ID %s: %w", v.ID, err)
		}

		menuTree := BuildMenuTree(allMenu, menu)
		myRoles = append(myRoles, auth.Role{
			ID:   v.ID,
			Name: v.Name,
			Menu: menuTree,
		})
	}

	meResponse := &auth.GetResponse{
		ID:       user.ID,
		Name:     user.Name,
		Username: "",
		Email:    user.Email,
		Phone:    "",
		Avatar:   "https://example.com/avatar.png",
		Role:     myRoles,
	}

	return meResponse, nil
}

func (u *authUsecase) Token(ctx context.Context, req *auth.TokenRequest) (*auth.TokenResponse, error) {
	db := u.DB.GDB.WithContext(ctx)

	user, err := u.UserRepository.GetByRefrehToken(ctx, db, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by refresh token: %w", err)
	}

	if user == nil {
		return nil, helper.Error(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	token, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (any, error) {
		return []byte(u.Config.GetString("jwt.refresh_secret")), nil
	})
	if err != nil || !token.Valid {
		u.Log.Error("failed to parse refresh token: ", err)
		return nil, helper.Error(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil || claims["token_id"] == nil {
		return nil, helper.Error(http.StatusUnauthorized, "invalid token claims")
	}

	userID := uuid.MustParse(claims["user_id"].(string))
	accessToken, err := GenerateAccessToken(userID, u.Config.GetString("jwt.secret"), time.Duration(u.Config.GetInt("jwt.access_token_expiry"))*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	user.AccessToken = accessToken
	if err := u.UserRepository.Update(ctx, db, user); err != nil {
		return nil, fmt.Errorf("failed to update user access token: %w", err)
	}

	if err := u.SetPermission(ctx, user.ID); err != nil {
		return nil, err
	}

	return auth.ToTokenResponse(accessToken, u.Config.GetInt("jwt.access_token_expiry")), nil
}

func (u *authUsecase) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	db := u.DB.GDB.WithContext(ctx)

	user, err := u.UserRepository.GetByEmail(ctx, db, req.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, helper.Error(http.StatusBadRequest, "invalid email or password")
	}

	if !isPasswordCorrect(req.Password, user.Password) {
		return nil, helper.Error(http.StatusBadRequest, "incorrect password")
	}

	accessToken, err := GenerateAccessToken(
		user.ID,
		u.Config.GetString("jwt.secret"),
		time.Duration(u.Config.GetInt("jwt.access_token_expiry"))*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := GenerateRefreshToken(
		user.ID,
		u.Config.GetString("jwt.refresh_secret"),
		time.Duration(u.Config.GetInt("jwt.refresh_token_expiry"))*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// key := fmt.Sprintf("refresh:%s:%s", user.ID, tokenID)
	// err = u.DB.Redis.Set(ctx, key, "valid", time.Duration(viper.GetInt("jwt.refresh_token_expiry"))*time.Second).Err()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to set refresh token in Redis: %w", err)
	// }

	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	if err := u.UserRepository.Update(ctx, db, user); err != nil {
		return nil, err
	}

	if err := u.SetPermission(ctx, user.ID); err != nil {
		return nil, err
	}

	return auth.ToLoginResponse(
		user,
		u.Config.GetInt("jwt.access_token_expiry"),
		u.Config.GetInt("jwt.refresh_token_expiry"),
	), nil
}

func (u *authUsecase) Logout(ctx context.Context, req *auth.LogoutRequest) error {
	db := u.DB.GDB.WithContext(ctx)

	user, err := u.UserRepository.GetByID(ctx, db, req.ID)
	if err != nil {
		return err
	}

	user.AccessToken = ""
	user.RefreshToken = ""
	if err := u.UserRepository.Update(ctx, db, user); err != nil {
		return err
	}

	return nil
}

func (u *authUsecase) SetPermission(ctx context.Context, id uuid.UUID) error {
	permissions, err := u.PermissionRepository.GetPermission(ctx, u.DB.GDB.WithContext(ctx), id)
	if err != nil {
		return fmt.Errorf("failed to get permissions for user %s: %w", id, err)
	}

	key := fmt.Sprintf("permissions:%s", id.String())
	value, _ := json.Marshal(permissions)
	if err := u.RedisRepository.Set(ctx, key, value, time.Duration(u.Config.GetInt("jwt.access_token_expiry"))*time.Second); err != nil {
		u.Log.Error("failed to set permissions in Redis: ", err)
		return fmt.Errorf("failed to set permissions in Redis: %w", err)
	}

	return nil
}

func (u *authUsecase) GetPermissions(ctx context.Context, id uuid.UUID) (model.Permission, error) {
	return u.PermissionRepository.GetPermission(ctx, u.DB.GDB, id)
}

func (u *authUsecase) GetPermissionFromRedis(ctx context.Context, key string) (map[string]struct{}, error) {
	permission, err := u.RedisRepository.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, helper.Error(http.StatusUnauthorized, "Unauthorized")
		}
		u.Log.Errorf("failed to get permission from redis: %v\n", err)
		return nil, fmt.Errorf("failed to get permission from redis: %w", err)
	}

	var permissions map[string]struct{}
	if err := json.Unmarshal([]byte(permission), &permissions); err != nil {
		u.Log.Errorf("failed to unmarshal permissions: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return permissions, nil
}

func BuildMenuTree(allMenu, filteredMenu []entity.Menu) []auth.Menu {
	// Index all menus by ID
	allMap := make(map[uuid.UUID]entity.Menu)
	for _, mnu := range allMenu {
		allMap[mnu.ID] = mnu
	}

	// Include all filtered + their ancestors
	includeMap := make(map[uuid.UUID]entity.Menu)
	var addAncestors func(mnu entity.Menu)
	addAncestors = func(mnu entity.Menu) {
		if _, exists := includeMap[mnu.ID]; exists {
			return
		}
		includeMap[mnu.ID] = mnu
		if mnu.ParentID != nil {
			if parent, found := allMap[*mnu.ParentID]; found {
				addAncestors(parent)
			}
		}
	}

	for _, mnu := range filteredMenu {
		addAncestors(mnu)
	}

	// Build GetAllResponse nodes
	nodeMap := make(map[uuid.UUID]*auth.Menu)
	for _, mnu := range includeMap {
		icon := ""
		if mnu.Icon != nil {
			icon = *mnu.Icon
		}
		order := 0
		if mnu.Order != nil {
			order = *mnu.Order
		}
		nodeMap[mnu.ID] = &auth.Menu{
			ID:       mnu.ID,
			Name:     mnu.Name,
			Path:     mnu.Path,
			Icon:     icon,
			Order:    order,
			IsActive: mnu.IsActive,
			Child:    []auth.Menu{},
		}
	}

	// Connect children to parents
	for _, mnu := range includeMap {
		node := nodeMap[mnu.ID]
		if mnu.ParentID != nil {
			if parentNode, ok := nodeMap[*mnu.ParentID]; ok {
				parentNode.Child = append(parentNode.Child, *node)
			}
		}
	}

	// Recursively sort children by Order
	var sortChildren func(nodes []auth.Menu)
	sortChildren = func(nodes []auth.Menu) {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Order < nodes[j].Order
		})
		for i := range nodes {
			sortChildren(nodes[i].Child)
		}
	}

	// Collect root nodes
	var roots []auth.Menu
	for _, node := range nodeMap {
		if original, ok := allMap[node.ID]; ok && original.ParentID == nil {
			roots = append(roots, *node)
		}
	}

	// Sort root nodes
	sortChildren(roots)

	return roots
}

func GenerateAccessToken(userID uuid.UUID, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uuid.UUID, secret string, duration time.Duration) (string, string, error) {
	tokenID := uuid.NewString()
	claims := jwt.MapClaims{
		"user_id":  userID,
		"token_id": tokenID,
		"exp":      time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	return signedToken, tokenID, err
}

func isPasswordCorrect(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
