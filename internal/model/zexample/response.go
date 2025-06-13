package zexample

import (
	"goilerplate/internal/entity"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetResponse struct {
	ID                 uuid.UUID        `json:"id"`
	BigIntNotNull      int64            `json:"bigintNotNull"`
	BigIntNull         *int64           `json:"bigintNull"`
	IntegerNotNull     int32            `json:"integerNotNull"`
	IntegerNull        *int32           `json:"integerNull"`
	NumericNotNull     decimal.Decimal  `json:"numericNotNull"`
	NumericNull        *decimal.Decimal `json:"numericNull"`
	VarcharNotNull     string           `json:"varcharNotNull"`
	VarcharNull        *string          `json:"varcharNull"`
	TextNotNull        string           `json:"textNotNull"`
	TextNull           *string          `json:"textNull"`
	BooleanNotNull     bool             `json:"booleanNotNull"`
	BooleanNull        *bool            `json:"booleanNull"`
	DateNotNull        time.Time        `json:"dateNotNull"`
	DateNull           *time.Time       `json:"dateNull"`
	TimestamptzNotNull time.Time        `json:"timestamptzNotNull"`
	TimestamptzNull    *time.Time       `json:"timestamptzNull"`
	TimestampNotNull   time.Time        `json:"timestampNotNull"`
	TimestampNull      *time.Time       `json:"timestampNull"`
	IsActive           bool             `json:"isActive"`
	UpdatedAt          *time.Time       `json:"updatedAt"`
	UpdatedBy          *uuid.UUID       `json:"updatedBy"`
	CustomeField1      string           `json:"customeField_1"`
	CustomeField2      string           `json:"customeField_2"`
	CustomeField3      string           `json:"customeField_3"`
	CustomeField4      string           `json:"customeField_4"`
	CustomeField5      string           `json:"customeField_5"`
}

type GetAllResponse struct {
	ID                 uuid.UUID        `json:"id"`
	BigIntNotNull      int64            `json:"bigintNotNull"`
	BigIntNull         *int64           `json:"bigintNull"`
	IntegerNotNull     int32            `json:"integerNotNull"`
	IntegerNull        *int32           `json:"integerNull"`
	NumericNotNull     decimal.Decimal  `json:"numericNotNull"`
	NumericNull        *decimal.Decimal `json:"numericNull"`
	VarcharNotNull     string           `json:"varcharNotNull"`
	VarcharNull        *string          `json:"varcharNull"`
	TextNotNull        string           `json:"textNotNull"`
	TextNull           *string          `json:"textNull"`
	BooleanNotNull     bool             `json:"booleanNotNull"`
	BooleanNull        *bool            `json:"booleanNull"`
	DateNotNull        time.Time        `json:"dateNotNull"`
	DateNull           *time.Time       `json:"dateNull"`
	TimestamptzNotNull time.Time        `json:"timestamptzNotNull"`
	TimestamptzNull    *time.Time       `json:"timestamptzNull"`
	TimestampNotNull   time.Time        `json:"timestampNotNull"`
	TimestampNull      *time.Time       `json:"timestampNull"`
	IsActive           bool             `json:"isActive"`
	UpdatedAt          *time.Time       `json:"updatedAt"`
	UpdatedBy          *uuid.UUID       `json:"updatedBy"`
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

func ToGetAll(example *entity.Example) *GetAllResponse {
	return &GetAllResponse{
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
