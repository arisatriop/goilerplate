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

type UpdateRequest struct {
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
