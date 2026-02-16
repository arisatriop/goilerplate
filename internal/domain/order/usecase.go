package order

import (
	"context"
	"fmt"
	"goilerplate/pkg/utils"
	"strings"
)

type Usecase interface {
	GenerateOrderNumber(orderType string) string
	GenerateQueueNumber(orderType string) string
	GetList(ctx context.Context, filter *Filter) ([]*Order, int64, error)
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) Usecase {
	return &usecase{
		repo: repo,
	}
}

func (u *usecase) GenerateOrderNumber(orderType string) string {
	timestamp := utils.Now()
	dmy := timestamp.Format("060102")
	unix := timestamp.Unix()

	ordType := "-"
	switch orderType {
	case OrderTypeDineIn:
		ordType = "DI"
	case OrderTypeTakeAway:
		ordType = "TA"
	}

	return strings.ToUpper(fmt.Sprintf("%s%s%d%s", ordType, dmy, unix, utils.GenerateRandomString(6)))
}

func (u *usecase) GenerateQueueNumber(orderType string) string {

	ordType := "-"
	switch orderType {
	case OrderTypeDineIn:
		ordType = "DI"
	case OrderTypeTakeAway:
		ordType = "TA"
	}

	// TODO: generate proper queue number: 4 digit human readable number. Ex: DI-0001, TA-0002
	return strings.ToUpper(fmt.Sprintf("%s-%s", ordType, utils.GenerateRandomNumberString(4)))
}

func (u *usecase) GetList(ctx context.Context, filter *Filter) ([]*Order, int64, error) {
	if filter == nil {
		filter = &Filter{}
	}

	orders, err := u.repo.GetListOrder(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.repo.CountOrders(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
