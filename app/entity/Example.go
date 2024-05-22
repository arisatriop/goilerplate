package entity

import "time"

type Example struct {
	Id        int64
	Code      string
	Example   string
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt time.Time
	UpdatedBy string
	DeletedAt time.Time
	DeletedBy string
	Uuid      string
}
