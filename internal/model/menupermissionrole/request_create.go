package menupermissionrole

type CreateRequest struct {
	MenuPermissionID string `json:"menuPermissionId" validate:"required,uuid"`
	RoleID           string `json:"roleId" validate:"required,uuid"`
}
