package redisearch

import (
	"testing"

	test "github.com/WALL-EEEEEEE/proxy-service/common/test"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {

	cases := []test.TestCase[any, any]{
		{
			Name:     "Search.AnyField",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{"*", "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewAnyField())
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TextField",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:"value1"`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewTextField("field1", NewStringValue("value1")))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TagField",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:{value1}`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewTagField("field1", NewStringValue("value1")))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.NumericField",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:[20,30]`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewNumericField("field1", NewNumericValue(20, 30)))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TagField.ValueOperatorAnd",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:{value1 value2}`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewTagField("field1", NewValueOperatorAnd(NewStringValue("value1"), NewStringValue("value2"))))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TextField.ValueOperatorAnd",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:"value1 value2"`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewTextField("field1", NewValueOperatorAnd(NewStringValue("value1"), NewStringValue("value2"))))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TagField.ValueOperatorOr",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:{value1|value2}`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewTagField("field1", NewValueOperatorOr(NewStringValue("value1"), NewStringValue("value2"))))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TextField.ValueOperatorOr",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:"value1|value2"`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewTextField("field1", NewValueOperatorOr(NewStringValue("value1"), NewStringValue("value2"))))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TagField.ValueOperatorNot",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:{-value1}`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewTagField("field1", NewValueOperatorNot(NewStringValue("value1"))))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TextField.ValueOperatorNot",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`@field1:"-value1"`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewTextField("field1", NewValueOperatorNot(NewStringValue("value1"))))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TextField.FieldOperatorAnd",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`(@field1:"value1" @field2:"value2")`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorAnd(
					NewTextField("field1", NewStringValue("value1")),
					NewTextField("field2", NewStringValue("value2")),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TagField.FieldOperatorAnd",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`(@field1:{value1} @field2:{value2})`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorAnd(
					NewTagField("field1", NewStringValue("value1")),
					NewTagField("field2", NewStringValue("value2")),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.NumericField.FieldOperatorAnd",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`(@field1:[20,30] @field2:[30,40])`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorAnd(
					NewNumericField("field1", NewNumericValue(20, 30)),
					NewNumericField("field2", NewNumericValue(30, 40)),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TextField.FieldOperatorOr",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`(@field1:"value1" | @field2:"value2")`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorOr(
					NewTextField("field1", NewStringValue("value1")),
					NewTextField("field2", NewStringValue("value2")),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TagField.FieldOperatorOr",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`(@field1:{value1} | @field2:{value2})`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorOr(
					NewTagField("field1", NewStringValue("value1")),
					NewTagField("field2", NewStringValue("value2")),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.NumericField.FieldOperatorOr",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`(@field1:[20,30] | @field2:[30,40])`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorOr(
					NewNumericField("field1", NewNumericValue(20, 30)),
					NewNumericField("field2", NewNumericValue(30, 40)),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TextField.FieldOperatorNot",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`-(@field1:"value1")`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorNot(
					NewTextField("field1", NewStringValue("value1")),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.TagField.FieldOperatorNot",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`-(@field1:{value1})`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorNot(
					NewTagField("field1", NewStringValue("value1")),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
		{
			Name:     "Search.NumericField.FieldOperatorNot",
			Input:    "",
			Error:    nil,
			Expected: []interface{}{`-(@field1:[20,30])`, "LIMIT", 10, 0},
			Check: func(tc test.TestCase[any, any]) {
				q := NewQuery(0, 10, NewFieldOperatorNot(
					NewNumericField("field1", NewNumericValue(20, 30)),
				))
				assert.Equal(t, q.Args(), tc.Expected)
			},
		},
	}
	test.Run(cases, t)
}
