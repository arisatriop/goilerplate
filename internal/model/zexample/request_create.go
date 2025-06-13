package zexample

import (
	"fmt"
	"goilerplate/internal/entity"
	"goilerplate/internal/helper"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateRequest struct {
	BigintNotNull int64  `json:"bigintNotNull" validate:"required"`
	BigintNull    *int64 `json:"bigintNull"`

	IntegerNotNull int32  `json:"integerNotNull" validate:"required"`
	IntegerNull    *int32 `json:"integerNull"`

	NumericNotNull string  `json:"numericNotNull" validate:"required"`
	NumericNull    *string `json:"numericNull"`

	VarcharNotNull string  `json:"varcharNotNull" validate:"required"`
	VarcharNull    *string `json:"varcharNull"`

	TextNotNull string  `json:"textNotNull" validate:"required"`
	TextNull    *string `json:"textNull"`

	BooleanNotNull bool  `json:"booleanNotNull" validate:""`
	BooleanNull    *bool `json:"booleanNull"`

	DateNotNull string  `json:"dateNotNull" validate:"required"`
	DateNull    *string `json:"dateNull"`

	TimestamptzNotNull time.Time  `json:"timestamptzNotNull" validate:"required"`
	TimestamptzNull    *time.Time `json:"timestamptzNull"`

	TimestampNotNull time.Time  `json:"timestampNotNull" validate:"required"`
	TimestampNull    *time.Time `json:"timestampNull"`

	CreatedBy uuid.UUID
}

func (r *CreateRequest) ToCreate() (*entity.Example, error) {
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
