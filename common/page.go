package common

import "encoding/json"

type Paginator[T any] struct {
	Total  int64
	Offset int64
	Count  int64
	Limit  int64
	Items  []T
}

func (page *Paginator[T]) MarshalBinary() ([]byte, error) {
	return json.Marshal(page)
}

func (page *Paginator[T]) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, page); err != nil {
		return err
	}
	return nil
}
func (page *Paginator[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(page)
}

func (page *Paginator[T]) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, page); err != nil {
		return err
	}
	return nil
}
