package checks

import (
	"context"
	"reflect"
)

const (
	base64URLAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
)

// //nolint: gochecknoglobals
var base64URLChars [256]bool

// //nolint: gochecknoinits
func init() {
	for i := range len(base64URLAlphabet) {
		base64URLChars[base64URLAlphabet[i]] = true
	}
}

func IsNil(val any) bool {
	if val == nil {
		return true
	}

	refVal := reflect.ValueOf(val)
	// //nolint: exhaustive
	switch refVal.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return refVal.IsNil()
	default:
		return false
	}
}

func IsPointer(data any) bool {
	return reflect.ValueOf(data).Kind() == reflect.Ptr
}

func Done[T any](ctx context.Context, ch <-chan T) bool {
	select {
	case <-ctx.Done():
		return true
	case <-ch:
		return true
	default:
	}

	return false
}

// CheckBase64URL checks whether or not s contains only Base64URL symbols.
func CheckBase64URL(s string) bool {
	for i := len(s) - 1; i >= 0; i-- {
		if !base64URLChars[s[i]] {
			return false
		}
	}

	return true
}
