// //nolint: revive
package types

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const (
	undef = "<undefined>"
)

// //nolint:exhaustive,mnd
func asString(src any) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}

	value := reflect.ValueOf(src)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(value.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	default:
		return fmt.Sprintf("%v", src)
	}
}

func strconvErr(err error) error {
	var ne *strconv.NumError

	if errors.As(err, &ne) {
		return ne.Err
	}

	return err
}
