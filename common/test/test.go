package test

import (
	"testing"
)

type TestCase[T any, E any] struct {
	Name     string
	Error    error
	Input    T
	Expected E
	Check    func(TestCase[T, E])
}

func Run(cases []TestCase[any, any], t *testing.T) {
	for _, t_case := range cases {
		t.Run(t_case.Name, func(t *testing.T) {
			t_case.Check(t_case)
		})
	}
}
