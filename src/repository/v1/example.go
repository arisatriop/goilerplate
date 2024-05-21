package v1

import (
	"goilerplate/api/request"
	"goilerplate/config"
	"goilerplate/src/entity"
)

type IExample interface {
	Create(con *config.Con, example *entity.Example) error
	Update(con *config.Con, id int64, example *entity.Example) error
	Delete(con *config.Con, id int64) error
	FindAll(con *config.Con, payload *request.ExampleReadPayload) ([]*entity.Example, error)
	FindById(con *config.Con, id int64) (*entity.Example, error)
}

type ExampleImpl struct{}

func NewExampleImpl() IExample {
	return &ExampleImpl{}
}

func (repository *ExampleImpl) Create(con *config.Con, example *entity.Example) error {
	panic("Not implement")
}

func (repository *ExampleImpl) Update(con *config.Con, id int64, example *entity.Example) error {
	panic("Not implement")
}

func (repository *ExampleImpl) Delete(con *config.Con, id int64) error {
	panic("Not implement")
}

func (repository *ExampleImpl) FindAll(con *config.Con, payload *request.ExampleReadPayload) ([]*entity.Example, error) {
	panic("Not implement")
}

func (repository *ExampleImpl) FindById(con *config.Con, id int64) (*entity.Example, error) {
	panic("Not implement")
}
