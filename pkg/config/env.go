// //nolint: revive
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	tagName = "env"
)

var ErrInvalidStruct = errors.New("invalid config struct")

var (
	splitCamelRegexp = regexp.MustCompile("([^A-Z]+|[A-Z]+[^A-Z]+|[A-Z]+)")
	acronymRegexp    = regexp.MustCompile("([A-Z]+)([A-Z][^A-Z]+)")
)

type cfgEntry struct {
	name  string
	alt   string
	key   string
	field reflect.Value
	tags  reflect.StructTag
}

//nolint:gocognit,mnd
func collect(prefix string, cfgStruct any) ([]cfgEntry, error) {
	structVal := reflect.ValueOf(cfgStruct)

	if structVal.Kind() != reflect.Ptr {
		return nil, ErrInvalidStruct
	}

	structVal = structVal.Elem()

	if structVal.Kind() != reflect.Struct {
		return nil, ErrInvalidStruct
	}

	typeOfEntry := structVal.Type()

	infos := make([]cfgEntry, 0, structVal.NumField())

	// iterate over struct fields
	for fieldIndex := range structVal.NumField() {
		field := structVal.Field(fieldIndex)

		if !field.CanSet() {
			continue
		}

		// could not process pointers
		if field.Kind() == reflect.Ptr {
			continue
		}

		ftype := typeOfEntry.Field(fieldIndex)
		entry := cfgEntry{
			name:  ftype.Name,
			field: field,
			tags:  ftype.Tag,
			alt:   strings.ToUpper(ftype.Tag.Get(tagName)),
		}

		entry.key = entry.name

		words := splitCamelRegexp.FindAllStringSubmatch(ftype.Name, -1)
		if len(words) > 0 {
			var name []string

			for _, words := range words {
				if m := acronymRegexp.FindStringSubmatch(words[0]); len(m) == 3 {
					name = append(name, m[1], m[2])
				} else {
					name = append(name, words[0])
				}
			}

			entry.key = strings.Join(name, "_")
		}

		if entry.alt != "" {
			entry.key = entry.alt
		}

		if prefix != "" {
			entry.key = fmt.Sprintf("%s_%s", prefix, entry.key)
		}

		entry.key = strings.ToUpper(entry.key)

		if field.Kind() == reflect.Struct {
			innerPrefix := prefix

			if !ftype.Anonymous {
				innerPrefix = entry.key
			}

			embeddedPtr := field.Addr().Interface()

			embeddedInfos, err := collect(innerPrefix, embeddedPtr)
			if err != nil {
				return nil, err
			}

			infos = append(infos, embeddedInfos...)
		} else {
			infos = append(infos, entry)
		}
	}

	return infos, nil
}

func InjectFromEnv(prefix string, cfgStruct any) error {
	entries, err := collect(prefix, cfgStruct)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		value, ok := os.LookupEnv(entry.key)
		if !ok && entry.alt != "" {
			value, ok = os.LookupEnv(entry.alt)
		}

		if !ok {
			continue
		}

		err = setValue(value, entry.field)
		if err != nil {
			return err
		}
	}

	return nil
}

//nolint:exhaustive
func setValue(value string, field reflect.Value) error {
	typ := field.Type()
	if typ.Kind() == reflect.Ptr {
		return nil
	}

	switch typ.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 0, typ.Bits())
		if err != nil {
			return err
		}

		field.SetUint(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}

		field.SetBool(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var val int64

		var err error

		if field.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(value)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, typ.Bits())
		}

		if err != nil {
			return err
		}

		field.SetInt(val)
	}

	return nil
}
