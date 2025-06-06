package zexample

import (
	"golang-clean-architecture/internal/entity"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetResponse struct {
	ID                 uuid.UUID        `json:"id"`
	BigIntNotNull      int64            `json:"bigint_not_null"`
	BigIntNull         *int64           `json:"bigint_null"`
	IntegerNotNull     int32            `json:"integer_not_null"`
	IntegerNull        *int32           `json:"integer_null"`
	NumericNotNull     decimal.Decimal  `json:"numeric_not_null"`
	NumericNull        *decimal.Decimal `json:"numeric_null"`
	VarcharNotNull     string           `json:"varchar_not_null"`
	VarcharNull        *string          `json:"varchar_null"`
	TextNotNull        string           `json:"text_not_null"`
	TextNull           *string          `json:"text_null"`
	BooleanNotNull     bool             `json:"boolean_not_null"`
	BooleanNull        *bool            `json:"boolean_null"`
	DateNotNull        time.Time        `json:"date_not_null"`
	DateNull           *time.Time       `json:"date_null"`
	TimestamptzNotNull time.Time        `json:"timestamptz_not_null"`
	TimestamptzNull    *time.Time       `json:"timestamptz_null"`
	TimestampNotNull   time.Time        `json:"timestamp_not_null"`
	TimestampNull      *time.Time       `json:"timestamp_null"`
	IsActive           bool             `json:"is_active"`
	UpdatedAt          *time.Time       `json:"updated_at"`
	UpdatedBy          *uuid.UUID       `json:"updated_by"`
}

func ToGet(example *entity.Example) *GetResponse {

	return &GetResponse{
		ID:                 example.ID,
		BigIntNotNull:      example.BigIntNotNull,
		BigIntNull:         example.BigIntNull,
		IntegerNotNull:     example.IntegerNotNull,
		IntegerNull:        example.IntegerNull,
		NumericNotNull:     example.NumericNotNull,
		NumericNull:        example.NumericNull,
		VarcharNotNull:     example.VarcharNotNull,
		VarcharNull:        example.VarcharNull,
		TextNotNull:        example.TextNotNull,
		TextNull:           example.TextNull,
		BooleanNotNull:     example.BooleanNotNull,
		BooleanNull:        example.BooleanNull,
		DateNotNull:        example.DateNotNull,
		DateNull:           example.DateNull,
		TimestamptzNotNull: example.TimestampTZNotNull,
		TimestamptzNull:    example.TimestampTZNull,
		TimestampNotNull:   example.TimestampNotNull,
		TimestampNull:      example.TimestampNull,
		IsActive:           example.IsActive,
		UpdatedAt:          example.UpdatedAt,
		UpdatedBy:          example.UpdatedBy,
	}
}
