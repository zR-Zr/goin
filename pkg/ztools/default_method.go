package ztools

import "reflect"

// GetOrDefault 三目运算
func GetOrDefault[T any](value T, defaultValue T) T {
	if strValue, ok := any(value).(string); ok && strValue == "" {
		return defaultValue
	}

	if reflect.ValueOf(value).IsZero() {
		return defaultValue
	}

	return value
}
