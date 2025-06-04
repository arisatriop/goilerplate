package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExampleCreateRequest struct {
	ID uuid.UUID `json:"id" validate:"required"` // UUID is NOT NULL (usually generated server-side)

	BigintNotNull int64  `json:"bigint_not_null" validate:"required"`
	BigintNull    *int64 `json:"bigint_null"`

	IntegerNotNull int32  `json:"integer_not_null" validate:"required"`
	IntegerNull    *int32 `json:"integer_null"`

	NumericNotNull decimal.Decimal  `json:"numeric_not_null" validate:"required"`
	NumericNull    *decimal.Decimal `json:"numeric_null"`

	VarcharNotNull string  `json:"varchar_not_null" validate:"required"`
	VarcharNull    *string `json:"varchar_null"`

	TextNotNull string  `json:"text_not_null" validate:"required"`
	TextNull    *string `json:"text_null"`

	BooleanNotNull bool  `json:"boolean_not_null" validate:"required"`
	BooleanNull    *bool `json:"boolean_null"`

	DateNotNull time.Time  `json:"date_not_null" validate:"required"`
	DateNull    *time.Time `json:"date_null"`

	TimestamptzNotNull time.Time  `json:"timestamptz_not_null" validate:"required"`
	TimestamptzNull    *time.Time `json:"timestamptz_null"`

	TimestampNotNull time.Time  `json:"timestamp_not_null" validate:"required"`
	TimestampNull    *time.Time `json:"timestamp_null"`

	CreatedBy string
}

// type ExampleUpdateRequest struct {
// 	Name      string `json:"name" validate:"max=100"`
// 	UpdatedBy string
// }

// type ExampleDeleteRequest struct {
// 	DeletedBy string
// }

// type ExampleGetRequest struct {
// 	OtherTableID string `json:"some_id"`
// 	Param        *Params
// }

// type ExampleResponse struct {
// 	ID        string `json:"id"`
// 	Name      string `json:"name"`
// 	CreatedBy string `json:"created_by"`
// 	CreatedAt string `json:"created_at"`
// 	UpdatedBy string `json:"updated_by"`
// 	UpdatedAt string `json:"updated_at"`
// 	DeletedBy string `json:"deleted_by,omitempty"`
// 	DeletedAt string `json:"deleted_at,omitempty"`
// }

// type ExampleListReponse struct {
// 	ID        string `json:"id"`
// 	Name      string `json:"name"`
// 	UpdatedBy string `json:"updated_by"`
// 	UpdatedAt string `json:"updated_at"`
// }

// func ToExampleResponse(entity *entity.Example) *ExampleResponse {
// 	return &ExampleResponse{
// 		// ID:        entity.ID,
// 		// Name:      entity.Name,
// 		// CreatedAt: entity.CreatedAt,
// 		// CreatedBy: entity.CreatedBy,
// 		// UpdatedAt: *entity.UpdatedAt,
// 		// UpdatedBy: entity.UpdatedBy,
// 		// DeletedBy: entity.DeletedBy,
// 		// DeletedAt: *entity.DeletedAt,
// 	}
// }

// func ToExampleListResponse(entity *entity.Example) ExampleListReponse {
// 	return ExampleListReponse{
// 		// ID:        entity.ID,
// 		// Name:      entity.Name,
// 		// UpdatedBy: entity.UpdatedBy,
// 		// UpdatedAt: *entity.UpdatedAt,
// 	}
// }
