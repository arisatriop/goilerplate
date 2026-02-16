package dtorequest

type CategoryCreateRequest struct {
	Name []string `json:"name" validate:"required,unique,dive,required" binding:"required,dive,required"`
}

type CategoryListRequest struct {
	Keyword string `query:"search"`
	StoreID string `query:"store_id"`
}
