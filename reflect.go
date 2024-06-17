package cloudy

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func GetFieldString(v interface{}, field string) string {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	return f.String()
}

func SetFieldString(v interface{}, field string, value string) {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	f.SetString(value)
}

func NewInstance(v interface{}) interface{} {
	myStructPtr := reflect.New(reflect.TypeOf(v))

	// Dereference the pointer to get the actual struct value
	rtn := myStructPtr.Elem().Interface()
	return rtn
}

func NewInstancePtr(v interface{}) interface{} {
	if reflect.TypeOf(v) == nil {
		return fmt.Errorf("the type specified is a pointer... it needs to be an actual object")
	}

	return reflect.New(reflect.TypeOf(v)).Interface()
}

func NewInstanceT[T any](v interface{}) (T, error) {
	var zero T
	if reflect.TypeOf(v) == nil {
		return zero, fmt.Errorf("the type specified is a pointer... it needs to be an actual object")
	}

	return reflect.New(reflect.TypeOf(v)).Interface().(T), nil
}

func NewT[T any]() (T, error) {
	var zero T
	if reflect.TypeOf(zero) == nil {
		return zero, fmt.Errorf("the type specified is a pointer... it needs to be an actual object")
	}

	return zero, nil
}

func IsPointer(item interface{}) bool {
	if reflect.ValueOf(item).Kind() == reflect.Ptr {
		return true
	}
	return false
}

func UnmarshallT[T any](data []byte) (*T, error) {
	var err error
	var model T

	if len(data) == 0 {
		return nil, nil
	}

	err = json.Unmarshal(data, &model)
	if err != nil {
		fmt.Println(err)
	}

	return &model, err
}
