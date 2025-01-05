package slice

import "slices"

func Contains[T comparable](v T, a []T) bool {
	return IndexOf(v, a) > -1
}

func IndexOf[T comparable](element T, data []T) int {
	return slices.Index(data, element)
}

// Deduplicate возвращает новый slice, в котором нет дубликатов. Порядок элементов сохраняется.
func Deduplicate[T comparable](values []T) []T {
	//nolint:mnd
	if len(values) < 2 {
		return slices.Clone(values)
	}

	m := make(map[T]struct{}, len(values))

	result := make([]T, 0, len(m))

	for _, v := range values {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = struct{}{}
		}
	}

	return result
}

// HasSubSlice проверяет входят ли все элементы requested в orig.
// Если requested пустой или nil, то функция возвращает true.
func HasSubSlice[T comparable](orig, requested []T) bool {
	m := make(map[T]struct{})

	for i := range orig {
		m[orig[i]] = struct{}{}
	}

	for i := range requested {
		if _, ok := m[requested[i]]; !ok {
			return false
		}
	}

	return true
}

func Append[T any](v T, arr ...T) []T {
	l := len(arr) + 1

	result := make([]T, 0, l)

	result = append(result, arr...)
	result = append(result, v)

	return result
}
