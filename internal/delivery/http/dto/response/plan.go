package dtoresponse

import "github.com/google/uuid"

type PlanTypeResponse struct {
	ID       uuid.UUID        `json:"id"`
	Code     string           `json:"code"`
	Name     string           `json:"name"`
	IsActive bool             `json:"isActive"`
	Plans    []PlanItemResponse `json:"plans"`
}

type PlanItemResponse struct {
	ID             uuid.UUID `json:"id"`
	DurationInDays int       `json:"durationInDays"`
	Price          string    `json:"price"`
	IsActive       bool      `json:"isActive"`
}
