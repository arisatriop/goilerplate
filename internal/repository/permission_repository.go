package repository

import (
	"context"
	"golang-clean-architecture/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type IPermissionRepository interface {
	GetPermission(ctx context.Context, db *pgxpool.Pool, id uuid.UUID) (model.Permission, error)
}

type PermissionRepository struct {
	Log *logrus.Logger
}

func NewPermissionRepository(log *logrus.Logger) IPermissionRepository {
	return &PermissionRepository{
		Log: log,
	}
}

func (r *PermissionRepository) GetPermission(ctx context.Context, db *pgxpool.Pool, id uuid.UUID) (model.Permission, error) {
	query := `
		SELECT mp."permission"
		FROM users u 
		JOIN role_users ru ON ru.user_id = u.id
		JOIN menu_permission_roles mpr ON mpr.role_id = ru.roles_id
		JOIN menu_permissions mp ON mp.id = mpr.menu_permission_id
		JOIN menus m ON m.id = mp.menu_id
		WHERE u.id = $1`

	rows, err := db.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("failed to execute query: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	permission := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			r.Log.Errorf("failed to scan row: %v\n", err)
			return nil, err
		}
		permission[name] = struct{}{}
	}

	return permission, nil
}
