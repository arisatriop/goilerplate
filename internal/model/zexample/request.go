package zexample

import (
	"fmt"
	"golang-clean-architecture/internal/entity"
	"golang-clean-architecture/internal/helper"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateRequest struct {
	BigintNotNull int64  `json:"bigint_not_null" validate:"required"`
	BigintNull    *int64 `json:"bigint_null"`

	IntegerNotNull int32  `json:"integer_not_null" validate:"required"`
	IntegerNull    *int32 `json:"integer_null"`

	NumericNotNull string  `json:"numeric_not_null" validate:"required"`
	NumericNull    *string `json:"numeric_null"`

	VarcharNotNull string  `json:"varchar_not_null" validate:"required"`
	VarcharNull    *string `json:"varchar_null"`

	TextNotNull string  `json:"text_not_null" validate:"required"`
	TextNull    *string `json:"text_null"`

	BooleanNotNull bool  `json:"boolean_not_null" validate:""`
	BooleanNull    *bool `json:"boolean_null"`

	DateNotNull string  `json:"date_not_null" validate:"required"`
	DateNull    *string `json:"date_null"`

	TimestamptzNotNull time.Time  `json:"timestamptz_not_null" validate:"required"`
	TimestamptzNull    *time.Time `json:"timestamptz_null"`

	TimestampNotNull time.Time  `json:"timestamp_not_null" validate:"required"`
	TimestampNull    *time.Time `json:"timestamp_null"`

	CreatedBy uuid.UUID
}

func (r *CreateRequest) ToEntity() (*entity.Example, error) {
	now := helper.NowJakarta()

	var numericNull *decimal.Decimal
	if r.NumericNull != nil && *r.NumericNull != "" {
		s, err := decimal.NewFromString(*r.NumericNull)
		if err != nil {
			return nil, helper.Error(http.StatusBadRequest, fmt.Sprintf("invalid numeric_null value: %s", *r.NumericNull))
		}
		numericNull = &s
	}

	numericNotNull, err := decimal.NewFromString(r.NumericNotNull)
	if err != nil {
		return nil, helper.Error(http.StatusBadRequest, fmt.Sprintf("invalid numeric_not_null value: %s", r.NumericNotNull))
	}

	var dateNull *time.Time
	if r.DateNull != nil && *r.DateNull != "" {
		d, err := time.Parse("2006-01-02", *r.DateNull)
		if err != nil {
			return nil, helper.Error(http.StatusBadRequest, fmt.Sprintf("invalid date_null value: %s", *r.DateNull))
		}
		dateNull = &d
	}

	dataNotNull, err := time.Parse("2006-01-02", r.DateNotNull)
	if err != nil {
		return nil, helper.Error(http.StatusBadRequest, fmt.Sprintf("invalid date_not_null value: %s", r.DateNotNull))
	}

	return &entity.Example{
		BigIntNull:         r.BigintNull,
		IntegerNull:        r.IntegerNull,
		NumericNull:        numericNull,
		VarcharNull:        r.VarcharNull,
		TextNull:           r.TextNull,
		BooleanNull:        r.BooleanNull,
		DateNull:           dateNull,
		TimestampTZNull:    r.TimestamptzNull,
		TimestampNull:      r.TimestampNull,
		BigIntNotNull:      r.BigintNotNull,
		IntegerNotNull:     r.IntegerNotNull,
		NumericNotNull:     numericNotNull,
		VarcharNotNull:     r.VarcharNotNull,
		TextNotNull:        r.TextNotNull,
		BooleanNotNull:     r.BooleanNotNull,
		DateNotNull:        dataNotNull,
		TimestampTZNotNull: r.TimestamptzNotNull,
		TimestampNotNull:   r.TimestampNotNull,
		IsActive:           true,
		CreatedAt:          now,
		CreatedBy:          r.CreatedBy,
		UpdatedAt:          &now,
		UpdatedBy:          &r.CreatedBy,
	}, nil
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
