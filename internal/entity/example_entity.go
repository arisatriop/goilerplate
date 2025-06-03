package entity

type Example struct {
	ID        string  `gorm:"column:id;primaryKey"`
	Name      string  `gorm:"column:name"`
	CreatedAt string  `gorm:"column:created_at"`
	UpdatedAt *string `gorm:"column:updated_at"`
	DeletedAt *string `gorm:"column:deleted_at"`
	CreatedBy string  `gorm:"column:created_by"`
	UpdatedBy string  `gorm:"column:updated_by"`
	DeletedBy string  `gorm:"column:deleted_by"`
}

func (c *Example) TableName() string {
	return "examples"
}
