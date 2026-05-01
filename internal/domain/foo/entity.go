package foo

type Foo struct {
	ID   string
	Code string
	Foo  string
}

func (e *Foo) Clone() *Foo {
	return &Foo{
		ID:   e.ID,
		Code: e.Code,
		Foo:  e.Foo,
	}
}
