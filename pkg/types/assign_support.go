// //nolint: goerr113,funlen,gocyclo,nestif,revive
package types

import (
	"errors"
	"fmt"
)

var (
	ErrConvert = errors.New("convert value error")
	ErrDecode  = errors.New("decode value error")
)

func unquoteIfQuoted(value any) (string, error) {
	var bytes []byte

	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return "", fmt.Errorf("could not convert value '%+v' to byte array of type '%T'",
			value, value)
	}

	// If the amount is quoted, strip the quotes
	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}

	return string(bytes), nil
}
