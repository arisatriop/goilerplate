package response

import "goilerplate/app/entity"

type Example struct {
	Uuid    string `json:"uuid"`
	Example string `json:"example"`
	Code    string `json:"code"`
}

type IExample interface {
	Create(example *entity.Example) *Example
	Update(example *entity.Example) *Example
	Delete(example *entity.Example) *Example
	FindAll(example []entity.Example) []Example
	FindById(example *entity.Example) *Example
}

type NewExampleImpl struct{}

func NewExampleExampleResponse() IExample {
	return &NewExampleImpl{}
}

func (r *NewExampleImpl) Create(example *entity.Example) *Example {
	panic("Not implement")
}

func (r *NewExampleImpl) Update(example *entity.Example) *Example {
	panic("Not implement")
}

func (r *NewExampleImpl) Delete(example *entity.Example) *Example {
	panic("Not implement")
}

func (r *NewExampleImpl) FindAll(example []entity.Example) []Example {
	panic("Not implement")
}

func (r *NewExampleImpl) FindById(example *entity.Example) *Example {
	panic("Not implement")
}
