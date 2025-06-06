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

type UpdateRequest struct {
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

	UpdatedBy uuid.UUID
}

func (r *UpdateRequest) ToUpdate(entity *entity.Example) error {
	now := helper.NowJakarta()

	var numericNull *decimal.Decimal
	if r.NumericNull != nil && *r.NumericNull != "" {
		s, err := decimal.NewFromString(*r.NumericNull)
		if err != nil {
			return helper.Error(http.StatusBadRequest, fmt.Sprintf("invalid numeric_null value: %s", *r.NumericNull))
		}
		numericNull = &s
	}

	numericNotNull, err := decimal.NewFromString(r.NumericNotNull)
	if err != nil {
		return helper.Error(http.StatusBadRequest, fmt.Sprintf("invalid numeric_not_null value: %s", r.NumericNotNull))
	}

	var dateNull *time.Time
	if r.DateNull != nil && *r.DateNull != "" {
		d, err := time.Parse("2006-01-02", *r.DateNull)
		if err != nil {
			return helper.Error(http.StatusBadRequest, fmt.Sprintf("invalid date_null value: %s", *r.DateNull))
		}
		dateNull = &d
	}

	dataNotNull, err := time.Parse("2006-01-02", r.DateNotNull)
	if err != nil {
		return helper.Error(http.StatusBadRequest, fmt.Sprintf("invalid date_not_null value: %s", r.DateNotNull))
	}

	entity.BigIntNull = r.BigintNull
	entity.IntegerNull = r.IntegerNull
	entity.NumericNull = numericNull
	entity.VarcharNull = r.VarcharNull
	entity.TextNull = r.TextNull
	entity.BooleanNull = r.BooleanNull
	entity.DateNull = dateNull
	entity.TimestampTZNull = r.TimestamptzNull
	entity.TimestampNull = r.TimestampNull
	entity.BigIntNotNull = r.BigintNotNull
	entity.IntegerNotNull = r.IntegerNotNull
	entity.NumericNotNull = numericNotNull
	entity.VarcharNotNull = r.VarcharNotNull
	entity.TextNotNull = r.TextNotNull
	entity.BooleanNotNull = r.BooleanNotNull
	entity.DateNotNull = dataNotNull
	entity.TimestampTZNotNull = r.TimestamptzNotNull
	entity.TimestampNotNull = r.TimestampNotNull
	entity.UpdatedAt = &now
	entity.UpdatedBy = &r.UpdatedBy

	return nil
}
