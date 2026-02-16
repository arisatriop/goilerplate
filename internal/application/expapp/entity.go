package expapp

import (
	"goilerplate/internal/domain/zexample"
	"goilerplate/internal/domain/zexample2"
)

type Exp struct {
	Example  *zexample.Example
	Example2 []*zexample2.Example2
}
