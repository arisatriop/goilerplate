package bar

type Usecase interface{}

type usecase struct{}

func NewUseCase() Usecase {
	return &usecase{}
}
