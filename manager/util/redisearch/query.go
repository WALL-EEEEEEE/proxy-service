package redisearch

import (
	"encoding/json"
	"fmt"
	"strings"
)

func escape(str string) string {
	characters := []string{
		".",
		"-",
		"@",
		"~",
		"|",
	}
	for _, c := range characters {
		str = strings.ReplaceAll(str, c, "\\"+c)
	}
	return str
}

type FieldKind string

const (
	FIELD_KIND_TEXT    FieldKind = "TEXT"
	FIELD_KIND_TAG     FieldKind = "TAG"
	FIELD_KIND_NUMERIC FieldKind = "NUMERIC"
)

type FieldValue[T any] interface {
	GetValue() T
	toQuery() string
}

type IStringValue = FieldValue[string]

type StringValue struct {
	value string
}

func (v StringValue) toQuery() string {
	return escape(v.value)
}

func (v StringValue) GetValue() string {
	return v.value
}

func NewStringValue(v string) StringValue {
	return StringValue{value: v}
}

type INumericValue = FieldValue[[2]int64]

type NumericValue struct {
	start int64
	end   int64
}

func (v NumericValue) toQuery() string {
	return fmt.Sprintf("[%d,%d]", v.start, v.end)
}

func (v NumericValue) GetValue() [2]int64 {
	return [2]int64{v.start, v.end}
}

func NewNumericValue(start int64, end int64) NumericValue {
	return NumericValue{start: start, end: end}
}

type ValueOperatorAnd[T any] struct {
	loperand FieldValue[T]
	roperand FieldValue[T]
}

func NewValueOperatorAnd[T any](l FieldValue[T], r FieldValue[T]) ValueOperatorAnd[T] {
	return ValueOperatorAnd[T]{loperand: l, roperand: r}
}

func (v ValueOperatorAnd[T]) toQuery() string {
	return fmt.Sprintf("%s %s", v.loperand.toQuery(), v.roperand.toQuery())
}

func (v ValueOperatorAnd[T]) GetValue() T {
	var rv T
	return rv
}

type ValueOperatorOr[T any] struct {
	loperand FieldValue[T]
	roperand FieldValue[T]
}

func (v ValueOperatorOr[T]) toQuery() string {
	return fmt.Sprintf("%s|%s", v.loperand.toQuery(), v.roperand.toQuery())
}

func (v ValueOperatorOr[T]) GetValue() T {
	var rv T
	return rv
}

func NewValueOperatorOr[T any](l FieldValue[T], r FieldValue[T]) ValueOperatorOr[T] {
	return ValueOperatorOr[T]{loperand: l, roperand: r}
}

type ValueOperatorNot[T any] struct {
	operand FieldValue[T]
}

func (v ValueOperatorNot[T]) toQuery() string {
	return fmt.Sprintf("-%s", v.operand.toQuery())
}

func (v ValueOperatorNot[T]) GetValue() T {
	var rv T
	return rv
}

func NewValueOperatorNot[T any](op FieldValue[T]) ValueOperatorNot[T] {
	return ValueOperatorNot[T]{operand: op}
}

type Field interface {
	toQuery() string
	GetName() string
}

type AnyField struct {
}

func (f AnyField) toQuery() string {
	return "*"
}

func (f AnyField) GetName() string {
	return ""
}
func NewAnyField() AnyField {
	return AnyField{}
}

type TextField struct {
	name  string
	value IStringValue
}

func (f TextField) toQuery() string {
	return fmt.Sprintf(`@%s:"%s"`, f.name, f.value.toQuery())
}

func (f TextField) GetName() string {
	return f.name
}
func NewTextField(name string, value IStringValue) TextField {
	return TextField{name: name, value: value}
}

type TagField struct {
	name  string
	value IStringValue
}

func (f TagField) toQuery() string {
	return fmt.Sprintf(`@%s:{%s}`, f.name, f.value.toQuery())
}
func (f TagField) GetName() string {
	return f.name
}

func NewTagField(name string, value IStringValue) TagField {
	return TagField{name: name, value: value}
}

type NumericField struct {
	name  string
	value NumericValue
}

func (f NumericField) toQuery() string {
	return fmt.Sprintf(`@%s:%s`, f.name, f.value.toQuery())
}
func (f NumericField) GetName() string {
	return f.name
}

func NewNumericField(name string, value NumericValue) NumericField {
	return NumericField{name: name, value: value}
}

type FieldOperatorAnd struct {
	loperand Field
	roperand Field
}

func (f FieldOperatorAnd) toQuery() string {
	return fmt.Sprintf("(%s %s)", f.loperand.toQuery(), f.roperand.toQuery())
}

func (f FieldOperatorAnd) GetName() string {
	return ""
}

func NewFieldOperatorAnd(l Field, r Field) FieldOperatorAnd {
	return FieldOperatorAnd{loperand: l, roperand: r}
}

type FieldOperatorOr struct {
	loperand Field
	roperand Field
}

func (f FieldOperatorOr) toQuery() string {
	return fmt.Sprintf("(%s)|(%s)", f.loperand.toQuery(), f.roperand.toQuery())
}

func (f FieldOperatorOr) GetName() string {
	return ""
}

func NewFieldOperatorOr(l Field, r Field) FieldOperatorOr {
	return FieldOperatorOr{loperand: l, roperand: r}
}

type FieldOperatorNot struct {
	operand Field
}

func (f FieldOperatorNot) toQuery() string {
	return fmt.Sprintf("-(%s)", f.operand.toQuery())

}

func (f FieldOperatorNot) GetName() string {
	return ""
}

func NewFieldOperatorNot(l Field) FieldOperatorNot {
	return FieldOperatorNot{operand: l}
}

type Query struct {
	field  Field
	limit  int
	offset int
}

func (q Query) Args() []interface{} {
	return []interface{}{q.field.toQuery(), "LIMIT", q.offset, q.limit}
}

func NewQuery(limit int, offset int, field Field) Query {
	return Query{limit: limit, offset: offset, field: field}
}

type SearchResultItem[T any] struct {
	Id   string
	Item T
}

type RawSearchResult interface{}

type SearchResult[T any] struct {
	Total int
	Items []SearchResultItem[T]
}

func ParseSearchResult[T any](raw_result RawSearchResult) (*SearchResult[T], error) {
	var result SearchResult[T] = SearchResult[T]{}
	_raw_result := raw_result.(map[interface{}]interface{})
	total := int(_raw_result["total_results"].(int64))
	result.Total = total
	raw_items := _raw_result["results"].([]interface{})
	var search_items []SearchResultItem[T] = make([]SearchResultItem[T], 0, total)
	for _, raw_item := range raw_items {
		_raw_item := raw_item.(map[interface{}]interface{})
		id := _raw_item["id"].(string)
		search_item := SearchResultItem[T]{Id: id}
		item := _raw_item["extra_attributes"].(map[interface{}]interface{})["$"].(string)
		json.Unmarshal([]byte(item), &search_item.Item)
		search_items = append(search_items, search_item)
	}
	result.Items = search_items
	return &result, nil
}
