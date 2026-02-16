package dtoresponse

type StoreResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Address     *string `json:"address"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	WebURL      string  `json:"webUrl"`
}
