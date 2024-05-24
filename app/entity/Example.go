package entity

import (
	"database/sql"
	"time"
)

type Example struct {
	Id        string
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
