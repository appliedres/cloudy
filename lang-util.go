package cloudy

import "strings"

func MapKeyStr(data map[string]interface{}, key string, caseInsensitive bool) (string, bool) {
	val, ok := data[key]

	if caseInsensitive && !ok {
		for k, v := range data {
			if strings.EqualFold(key, k) {
				val = v
				ok = true
				break
			}
		}

	}

	strVal, okConv := val.(string)
	if !okConv {
		return "", false
	}

	return strVal, ok
}

func MapKey[T any](data map[string]T, key string, caseInsensitive bool) (T, bool) {
	val, ok := data[key]

	if caseInsensitive && !ok {
		for k, v := range data {
			if strings.EqualFold(key, k) {
				val = v
				ok = true
				break
			}
		}
	}
	return val, ok
}

func StringP(s string) *string {
	return &s
}

func BoolP(v bool) *bool {
	return &v
}
