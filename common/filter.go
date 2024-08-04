package common

import (
	"encoding/json"

	"github.com/bobg/go-generics/maps"
)

type FilterOperator struct {
	Enum[string]
	Value string
}

func (filter_op *FilterOperator) String() string {
	return filter_op.Value
}

func (filter_op *FilterOperator) Values() []string {
	return maps.Keys(filter_operators)
}

func (filter_op *FilterOperator) Index() {
}

func (filter_op FilterOperator) Exists() bool {
	_, ok := filter_operators[filter_op.Value]
	return ok
}

func (filter_op *FilterOperator) MarshalJSON() ([]byte, error) {
	return json.Marshal(filter_op.Value)
}

func (filter_op *FilterOperator) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	filter_op.Value = value
	return nil
}

func (filter_op *FilterOperator) MarshalBinary() ([]byte, error) {
	return []byte(filter_op.Value), nil
}

func (filter_op *FilterOperator) UnmarshalBinary(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	filter_op.Value = value
	return nil
}

var (
	FILTER_LESS_THAN             FilterOperator         = FilterOperator{Value: "LESS_THAN"}
	FILTER_LESS_THAN_OR_EQUAL    FilterOperator         = FilterOperator{Value: "LESS_THAN_OR_EQUAL"}
	FILTER_GREATER_THAN          FilterOperator         = FilterOperator{Value: "GREATER_THAN"}
	FILTER_GREATER_THAN_OR_EQUAL FilterOperator         = FilterOperator{Value: "GREATER_THAN_OR_EQUAL"}
	FILTER_EQUAL                 FilterOperator         = FilterOperator{Value: "EQUAL"}
	FILTER_NOT_EQUAL             FilterOperator         = FilterOperator{Value: "NOT_EQUAL"}
	filter_operators             map[string]interface{} = map[string]interface{}{
		FILTER_LESS_THAN.String():             nil,
		FILTER_LESS_THAN_OR_EQUAL.String():    nil,
		FILTER_GREATER_THAN.String():          nil,
		FILTER_GREATER_THAN_OR_EQUAL.String(): nil,
		FILTER_EQUAL.String():                 nil,
		FILTER_NOT_EQUAL.String():             nil,
	}
)

type FilterSetOperator struct {
	Enum[string]
	Value string
}

func (filterset_op *FilterSetOperator) String() string {
	return filterset_op.Value
}

func (filterset_op *FilterSetOperator) Values() []string {
	return maps.Keys(filter_operators)
}

func (filterset_op *FilterSetOperator) Index() {
}

func (filterset_op FilterSetOperator) Exists() bool {
	_, ok := filterset_operators[filterset_op.Value]
	return ok
}
func (filterset_op *FilterSetOperator) MarshalBinary() ([]byte, error) {
	return []byte(filterset_op.Value), nil
}

func (filterset_op *FilterSetOperator) UnmarshalBinary(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	filterset_op.Value = value
	return nil
}
func (filterset_op *FilterSetOperator) MarshalJSON() ([]byte, error) {
	return json.Marshal(filterset_op.Value)
}

func (filterset_op *FilterSetOperator) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	filterset_op.Value = value
	return nil
}

var (
	FILTERSET_OR        FilterSetOperator      = FilterSetOperator{Value: "OR"}
	FILTERSET_AND       FilterSetOperator      = FilterSetOperator{Value: "AND"}
	filterset_operators map[string]interface{} = map[string]interface{}{
		FILTERSET_AND.String(): nil,
		FILTERSET_OR.String():  nil,
	}
)

type Filter[T any] struct {
	Name  string             `json:"name"`
	Op    *FilterOperator    `json:"op"`
	SetOp *FilterSetOperator `json:"set_op"`
	Value T                  `json:"value"`
}
