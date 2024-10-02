package cloudy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrDriverNotFound = errors.New("driver not found")
var ErrInvalidConfiguration = errors.New("invalid configuration object")
var ErrOperationNotImplemented = errors.New("operation not implemented")

type ProviderFactory2[T any] interface {
	New(ctx context.Context, cfg interface{}) (T, error)
	NewConfig() interface{}
}

type ProviderFactory[T any] interface {
	Create(cfg interface{}) (T, error)
	FromEnv(env *Environment) (interface{}, error)
}

type ProvidersRegistry[T any] struct {
	Providers map[string]ProviderFactory[T]
}

func NewProviderRegistry[T any]() *ProvidersRegistry[T] {
	return &ProvidersRegistry[T]{
		Providers: make(map[string]ProviderFactory[T]),
	}
}

func (pr *ProvidersRegistry[T]) Register(name string, factory ProviderFactory[T]) {
	// Info(context.Background(), "Registring Environment Provider: %s", name)
	pr.Providers[name] = factory
}

func (pr *ProvidersRegistry[T]) New(name string, cfg interface{}) (T, error) {
	var zero T
	factory, ok := MapKey(pr.Providers, name, true)
	if !ok {
		return zero, ErrDriverNotFound
	}

	return factory.Create(cfg)
}

func (pr *ProvidersRegistry[T]) NewFromEnv(env *Environment, driverKey string) (T, error) {
	var zero T

	driver := env.Force(driverKey, "DRIVER", "DEFAULT_DRIVER")
	factory, ok := MapKey(pr.Providers, driver, true)
	if !ok {
		var keys []string
		for key := range pr.Providers {
			keys = append(keys, key)
		}
		Error(context.Background(), "Driver Not found. Available drivers are: %v", keys)
		return zero, ErrDriverNotFound
	}
	cfg, err := factory.FromEnv(env)
	if err != nil {
		return zero, err
	}
	return factory.Create(cfg)
}

func (pr *ProvidersRegistry[T]) NewFromEnvWith(env *Environment, driver string) (T, error) {
	var zero T
	factory, ok := MapKey(pr.Providers, driver, true)
	if !ok {
		var keys []string
		for key := range pr.Providers {
			keys = append(keys, key)
		}
		Error(context.Background(), "Driver Not found. Available drivers are: %v", keys)
		return zero, ErrDriverNotFound
	}
	cfg, err := factory.FromEnv(env)
	if err != nil {
		return zero, err
	}
	return factory.Create(cfg)
}

func FromEnv(prefix string, driverKey string) (string, map[string]interface{}, error) {
	env := LoadEnvPrefixMap(prefix)
	driver, found := MapKeyStr(env, driverKey, true)

	if !found {
		return "", nil, fmt.Errorf("no driver found under key %v", driverKey)
	}
	if driver == "" {
		return "", nil, fmt.Errorf("empty driver found under key %v", driverKey)
	}

	return driver, env, nil
}

func (pr *ProvidersRegistry[T]) Print() {
	for n := range pr.Providers {
		fmt.Printf("- %v\n", n)
	}
}

type Storage interface {
}

// Loads all the environment variables with the given prefix.
// This normalizes the env variable names to all uppercase
// without underscores. This means that POOL_ID == POOLID == PoolId == pool_id
func LoadEnvPrefixMap(prefix string) map[string]interface{} {
	all := os.Environ()
	uprefix := strings.ToUpper(prefix)
	if !strings.HasPrefix(uprefix, "_") {
		uprefix += "_"
	}

	rtn := make(map[string]interface{})
	for _, env := range all {
		if env[0:1] == "\"" {
			env = env[1:]
		}
		if strings.HasSuffix(env, "\"") {
			env = env[:(len(env) - 1)]
		}
		i := strings.Index(env, "=")
		if i < 0 {
			continue
		}
		envName := env[0:i]
		envValue := env[i+1:]

		// Replace all the _ characters
		envu := strings.ToUpper(envName)

		if strings.HasPrefix(envu, uprefix) {
			key := envu[len(uprefix):]
			rtn[key] = envValue
		}
	}

	return rtn
}
