package entity

import (
	"database/sql"
	"time"
)

type Example struct {
	Id        int64
	Code      string
	Example   string
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt sql.NullTime
	UpdatedBy sql.NullString
	DeletedAt sql.NullTime
	DeletedBy sql.NullString
	Uuid      string
}
