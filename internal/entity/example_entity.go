package entity

import (
	"time"

	"github.com/google/uuid"
)

type Example struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	// Nullable fields as pointers for GORM compatibility,
	// pgxpool can scan into pointers as well.
	BigIntNull      *int64     `gorm:"column:bigint_null" json:"bigint_null"`
	IntegerNull     *int32     `gorm:"column:integer_null" json:"integer_null"`
	NumericNull     *float64   `gorm:"column:numeric_null" json:"numeric_null"`
	VarcharNull     *string    `gorm:"column:varchar_null" json:"varchar_null"`
	TextNull        *string    `gorm:"column:text_null" json:"text_null"`
	BooleanNull     *bool      `gorm:"column:boolean_null" json:"boolean_null"`
	DateNull        *time.Time `gorm:"column:date_null" json:"date_null"`
	TimestampTZNull *time.Time `gorm:"column:timestamptz_null" json:"timestamptz_null"`
	TimestampNull   *time.Time `gorm:"column:timestamp_null" json:"timestamp_null"`

	// NOT NULL fields as regular types
	BigIntNotNull      int64     `gorm:"column:bigint_not_null" json:"bigint_not_null"`
	IntegerNotNull     int32     `gorm:"column:integer_not_null" json:"integer_not_null"`
	NumericNotNull     float64   `gorm:"column:numeric_not_null" json:"numeric_not_null"`
	VarcharNotNull     string    `gorm:"column:varchar_not_null" json:"varchar_not_null"`
	TextNotNull        string    `gorm:"column:text_not_null" json:"text_not_null"`
	BooleanNotNull     bool      `gorm:"column:boolean_not_null" json:"boolean_not_null"`
	DateNotNull        time.Time `gorm:"column:date_not_null" json:"date_not_null"`
	TimestampTZNotNull time.Time `gorm:"column:timestamptz_not_null" json:"timestamptz_not_null"`
	TimestampNotNull   time.Time `gorm:"column:timestamp_not_null" json:"timestamp_not_null"`

	// Deafult column I recommend
	IsActive  bool       `gorm:"column:is_active"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedBy uuid.UUID  `gorm:"column:created_by"`
	UpdatedBy *uuid.UUID `gorm:"column:updated_by"`
	DeletedBy *uuid.UUID `gorm:"column:deleted_by"`
}

func (c *Example) TableName() string {
	return "examples"
}
