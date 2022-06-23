package cloudy

import (
	"os"
	"strconv"
	"strings"
)

//MapKeyStr is used for dealing with JSON...
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

func StringFromP(s *string, missing string) string {
	if s == nil {
		return missing
	}
	return *s
}

func BoolFromP(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IntEnv(name string) int {
	val := os.Getenv(name)
	if val == "" {
		return 0
	}
	num, _ := strconv.Atoi(val)
	return num
}

func IntP(v int) *int {
	return &v
}

func TrimDomain(v string) string {
	i := strings.Index(v, "@")
	if i > 0 {
		return v[:i]
	}
	return v
}

func RemoveDomain(email string) string {
	index := strings.Index(email, "@")
	if index > 0 {
		email = email[0:index]
	}
	return email
}

func StrContains(str string, arr []string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
