package model

import "golang-clean-architecture/internal/entity"

type ExampleCreateRequest struct {
	Name      string `json:"name" validate:"required,max=100"`
	CreatedBy string
}

type ExampleUpdateRequest struct {
	Name      string `json:"name" validate:"max=100"`
	UpdatedBy string
}

type ExampleDeleteRequest struct {
	DeletedBy string
}

type ExampleGetRequest struct {
	Keyword  string `json:"keyword"`
	SomeID   string `json:"some_id"`
	Paginate *RequestPagination
}

type ExampleResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedBy string `json:"updated_by"`
	UpdatedAt string `json:"updated_at"`
	DeletedBy string `json:"deleted_by,omitempty"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

type ExampleListReponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UpdatedBy string `json:"updated_by"`
	UpdatedAt string `json:"updated_at"`
}

func ToExampleResponse(entity *entity.Example) *ExampleResponse {
	return &ExampleResponse{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		CreatedBy: entity.CreatedBy,
		UpdatedAt: *entity.UpdatedAt,
		UpdatedBy: entity.UpdatedBy,
		DeletedBy: entity.DeletedBy,
		DeletedAt: *entity.DeletedAt,
	}
}

func ToExampleListResponse(entity *entity.Example) ExampleListReponse {
	return ExampleListReponse{
		ID:        entity.ID,
		Name:      entity.Name,
		UpdatedBy: entity.UpdatedBy,
		UpdatedAt: *entity.UpdatedAt,
	}
}
