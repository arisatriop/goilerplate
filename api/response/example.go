package response

import "goilerplate/app/entity"

type Example struct {
	Uuid    string `json:"uuid"`
	Code    string `json:"code"`
	Example string `json:"example"`
}

type IExample interface {
	Create(example *entity.Example) *Example
	Update(example *entity.Example) *Example
	Delete(example *entity.Example) *Example
	FindAll(example []*entity.Example) ([]*Example, error)
	FindById(example *entity.Example) (*Example, error)
}

type NewExampleImpl struct{}

func NewExampleResponse() IExample {
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

func (r *NewExampleImpl) FindAll(example []*entity.Example) ([]*Example, error) {
	var exps []*Example
	for _, e := range example {
		exps = append(exps, &Example{
			Uuid:    e.Uuid,
			Code:    e.Code,
			Example: e.Example,
		})
	}
	return exps, nil
}

func (r *NewExampleImpl) FindById(example *entity.Example) (*Example, error) {
	return &Example{
		Uuid:    example.Uuid,
		Code:    example.Code,
		Example: example.Example,
	}, nil
}
