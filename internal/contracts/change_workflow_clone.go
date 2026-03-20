package contracts

import (
	"reflect"
	"slices"
	"strings"
)

func duplicateString(items []string) (string, bool) {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			return item, true
		}
		seen[item] = struct{}{}
	}
	return "", false
}

func uniqueSortedStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	slices.Sort(result)
	return result
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func stringSliceToAny(items []string) []any {
	result := make([]any, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}

func requiredStringValue(raw any) string {
	value, _ := raw.(string)
	return value
}

func optionalStringValue(raw any) string {
	value, _ := raw.(string)
	return strings.TrimSpace(value)
}

func cloneTopLevelValue(value any) any {
	if value == nil {
		return nil
	}
	cloned := cloneReflectValue(reflect.ValueOf(value))
	if !cloned.IsValid() {
		return nil
	}
	return cloned.Interface()
}

func cloneReflectValue(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}
	switch value.Kind() {
	case reflect.Interface:
		return cloneInterfaceValue(value)
	case reflect.Slice:
		return cloneSliceValue(value)
	case reflect.Array:
		return cloneArrayValue(value)
	case reflect.Map:
		return cloneMapValue(value)
	case reflect.Pointer:
		return clonePointerValue(value)
	default:
		return value
	}
}

func cloneInterfaceValue(value reflect.Value) reflect.Value {
	if value.IsNil() {
		return reflect.Zero(value.Type())
	}
	cloned := cloneReflectValue(value.Elem())
	wrapped := reflect.New(cloned.Type()).Elem()
	wrapped.Set(cloned)
	return wrapped
}

func cloneSliceValue(value reflect.Value) reflect.Value {
	if value.IsNil() {
		return reflect.Zero(value.Type())
	}
	result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
	for i := 0; i < value.Len(); i++ {
		result.Index(i).Set(cloneReflectValue(value.Index(i)))
	}
	return result
}

func cloneArrayValue(value reflect.Value) reflect.Value {
	result := reflect.New(value.Type()).Elem()
	for i := 0; i < value.Len(); i++ {
		result.Index(i).Set(cloneReflectValue(value.Index(i)))
	}
	return result
}

func cloneMapValue(value reflect.Value) reflect.Value {
	if value.IsNil() {
		return reflect.Zero(value.Type())
	}
	result := reflect.MakeMapWithSize(value.Type(), value.Len())
	iter := value.MapRange()
	for iter.Next() {
		result.SetMapIndex(iter.Key(), cloneReflectValue(iter.Value()))
	}
	return result
}

func clonePointerValue(value reflect.Value) reflect.Value {
	if value.IsNil() {
		return reflect.Zero(value.Type())
	}
	result := reflect.New(value.Elem().Type())
	result.Elem().Set(cloneReflectValue(value.Elem()))
	return result
}
