package v1

import (
	"goilerplate/api/request"
	"goilerplate/app/entity"
	"goilerplate/config"
)

type IExample interface {
	Create(example *entity.Example) error
	Update(id int64, example *entity.Example) error
	Delete(id int64) error
	FindAll(payload *request.ExampleReadPayload) ([]*entity.Example, error)
	FindById(id int64) (*entity.Example, error)
}

type ExampleImpl struct {
	Con *config.Con
}

func NewExampleImpl(con *config.Con) IExample {
	return &ExampleImpl{
		Con: con,
	}
}

func (r *ExampleImpl) Create(example *entity.Example) error {
	panic("Not implement")
}

func (r *ExampleImpl) Update(id int64, example *entity.Example) error {
	panic("Not implement")
}

func (r *ExampleImpl) Delete(id int64) error {
	panic("Not implement")
}

func (r *ExampleImpl) FindAll(payload *request.ExampleReadPayload) ([]*entity.Example, error) {
	panic("Not implement")
}

func (r *ExampleImpl) FindById(id int64) (*entity.Example, error) {
	panic("Not implement")
}