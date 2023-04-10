package log

import "go.uber.org/zap"

type Field struct {
	field zap.Field
}

func (f Field) GetKey() string {
	return f.field.Key
}

func String(key, val string) Field {
	return Field{field: zap.String(key, val)}
}

func Int(key string, val int) Field {
	return Field{field: zap.Int(key, val)}
}

func Reflect(key string, val interface{}) Field {
	return Field{field: zap.Reflect(key, val)}
}

func NamedError(key string, err error) Field {
	return Field{field: zap.NamedError(key, err)}
}

func Error(err error) Field {
	return NamedError("error", err)
}
