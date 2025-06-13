package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Example struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	// Nullable fields as pointers for GORM compatibility,
	// pgxpool can scan into pointers as well.
	BigIntNull      *int64           `gorm:"column:bigint_null" json:"bigintNull"`
	IntegerNull     *int32           `gorm:"column:integer_null" json:"integerNull"`
	NumericNull     *decimal.Decimal `gorm:"column:numeric_null" json:"numericNull"`
	VarcharNull     *string          `gorm:"column:varchar_null" json:"varcharNull"`
	TextNull        *string          `gorm:"column:text_null" json:"textNull"`
	BooleanNull     *bool            `gorm:"column:boolean_null" json:"booleanNull"`
	DateNull        *time.Time       `gorm:"column:date_null" json:"dateNull"`
	TimestampTZNull *time.Time       `gorm:"column:timestamptz_null" json:"timestamptzNull"`
	TimestampNull   *time.Time       `gorm:"column:timestamp_null" json:"timestampNull"`

	// NOT NULL fields as regular types
	BigIntNotNull      int64           `gorm:"column:bigint_not_null" json:"bigintNotNull"`
	IntegerNotNull     int32           `gorm:"column:integer_not_null" json:"integerNotNull"`
	NumericNotNull     decimal.Decimal `gorm:"column:numeric_not_null" json:"numericNotNull"`
	VarcharNotNull     string          `gorm:"column:varchar_not_null" json:"varcharNotNull"`
	TextNotNull        string          `gorm:"column:text_not_null" json:"textNotNull"`
	BooleanNotNull     bool            `gorm:"column:boolean_not_null" json:"booleanNotNull"`
	DateNotNull        time.Time       `gorm:"column:date_not_null" json:"dateNotNull"`
	TimestampTZNotNull time.Time       `gorm:"column:timestamptz_not_null" json:"timestamptzNotNull"`
	TimestampNotNull   time.Time       `gorm:"column:timestamp_not_null" json:"timestampNotNull"`

	// Deafult column I recommend
	IsActive  bool       `gorm:"column:is_active"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedBy uuid.UUID  `gorm:"column:created_by"`
	UpdatedBy *uuid.UUID `gorm:"column:updated_by"`
	DeletedBy *uuid.UUID `gorm:"column:deleted_by"`
}

func (e *Example) TableName() string {
	return "examples"
}
