package cloudy

import (
	"errors"
)

var DriverNotFoundError = errors.New("driver not found")
var InvalidConfigurationError = errors.New("invalid configuration object")
var OperationNotImplementedError = errors.New("operation not implemented")

type ProviderFactory[T any] interface {
	Create(cfg interface{}) (T, error)
}

type ProvidersRegistry[T any] struct {
	Providers map[string]func(cfg interface{}) (T, error)
}

func NewProviderRegistry[T any]() *ProvidersRegistry[T] {
	return &ProvidersRegistry[T]{
		Providers: make(map[string]func(cfg interface{}) (T, error)),
	}
}

func (pr *ProvidersRegistry[T]) Register(name string, fn func(cfg interface{}) (T, error)) {
	pr.Providers[name] = fn
}

func (pr *ProvidersRegistry[T]) New(name string, cfg interface{}) (T, error) {
	var zero T
	fn, ok := MapKey(pr.Providers, name, true)
	if !ok {
		return zero, DriverNotFoundError
	}
	return fn(cfg)
}

type Storage interface {
}
