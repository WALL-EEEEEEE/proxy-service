package common

type Enum[T any] interface {
	String() string
	Index() T
	Values() []T
	Exists(T) bool
}
