package productcategory

import "github.com/google/uuid"

type ProductCategory struct {
	ProductID  uuid.UUID
	CategoryID uuid.UUID
	IsActive   bool
}
