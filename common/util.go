package common

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"strings"

	"github.com/bobg/go-generics/slices"
	"github.com/fatih/structs"
	fieldmask_utils "github.com/mennanov/fieldmask-utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GenerateUidByStrs(strs ...string) (*string, error) {
	concat_str := strings.Join(strs, ":")
	id_gen := md5.New()
	_, err := id_gen.Write([]byte(concat_str))
	if err != nil {
		return nil, err
	}
	uid := fmt.Sprintf("%x", id_gen.Sum(nil))
	return &uid, nil
}

func MaskNaming(name string) string {
	parts, _ := slices.Map(strings.Split(name, "_"), func(index int, value string) (string, error) {
		return cases.Title(language.English).String(value), nil
	})
	name = strings.Join(parts, "")
	return name
}

func MaskFields[T any](src *T, paths []string) (*T, error) {
	var target T
	_mask, err := fieldmask_utils.MaskFromPaths(paths, MaskNaming)
	if err != nil {
		return nil, err
	}
	err = fieldmask_utils.StructToStruct(_mask, src, &target)
	if err != nil {
		return nil, err
	}
	return &target, nil
}

// Patch updates the target struct in-place with non-zero values from the patch struct.
// Only fields with the same name and type get updated. Fields in the patch struct can be
// pointers to the target's type.
//
// Returns true if any value has been changed.
func Patch[T any](target, patch T) (changed bool, err error) {

	var dst = structs.New(target)
	var fields = structs.New(patch).Fields() // work stack

	for N := len(fields); N > 0; N = len(fields) {
		var srcField = fields[N-1] // pop the top
		fields = fields[:N-1]

		if !srcField.IsExported() {
			continue // skip unexported fields
		}
		if srcField.IsEmbedded() {
			// add the embedded fields into the work stack
			fields = append(fields, srcField.Fields()...)
			continue
		}
		if srcField.IsZero() {
			continue // skip zero-value fields
		}

		var name = srcField.Name()

		var dstField, ok = dst.FieldOk(name)
		if !ok {
			continue // skip non-existing fields
		}
		var srcValue = reflect.ValueOf(srcField.Value())
		srcValue = reflect.Indirect(srcValue)
		if skind, dkind := srcValue.Kind(), dstField.Kind(); skind != dkind {
			err = fmt.Errorf("field `%v` types mismatch while patching: %v vs %v", name, dkind, skind)
			return
		}

		if !reflect.DeepEqual(srcValue.Interface(), dstField.Value()) {
			changed = true
		}

		err = dstField.Set(srcValue.Interface())
		if err != nil {
			return
		}
	}
	return
}
