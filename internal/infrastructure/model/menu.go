package model

import "time"

// Menu represents the menus table
type Menu struct {
	ID           string     `gorm:"column:id;type:uuid;primaryKey"`
	ParentID     *string    `gorm:"column:parent_id;type:uuid"`
	Name         string     `gorm:"column:name;type:varchar(100);not null"`
	Slug         string     `gorm:"column:slug;type:varchar(100);unique;not null"`
	Icon         *string    `gorm:"column:icon;type:varchar(100)"`
	Route        *string    `gorm:"column:route;type:varchar(255)"`
	DisplayOrder float64    `gorm:"column:display_order;type:decimal(10,2);default:0"`
	IsActive     bool       `gorm:"column:is_active;type:boolean;default:true"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	CreatedBy    string     `gorm:"column:created_by;type:varchar(255);not null"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null"`
	UpdatedBy    string     `gorm:"column:updated_by;type:varchar(255);not null"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	DeletedBy    *string    `gorm:"column:deleted_by;type:varchar(255)"`
}

// TableName overrides the table name
func (Menu) TableName() string {
	return "menus"
}
