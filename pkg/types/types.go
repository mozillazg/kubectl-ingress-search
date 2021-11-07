package types

import "fmt"

type Value struct {
	V        fmt.Stringer
	Rendered string
}

type String string

func (v Value) Render() string {
	if v.Rendered != "" {
		return v.Rendered
	}
	return v.V.String()
}

func (s String) String() string {
	return string(s)
}

func (s String) ToValue() Value {
	return Value{
		V: s,
	}
}
