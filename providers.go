package cloudy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

var ErrDriverNotFound = errors.New("driver not found")
var ErrInvalidConfiguration = errors.New("invalid configuration object")
var ErrOperationNotImplemented = errors.New("operation not implemented")

type ProviderFactory[T any] interface {
	Create(cfg interface{}) (T, error)
	ToConfig(config map[string]interface{}) (interface{}, error)
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

func (pr *ProvidersRegistry[T]) NewFromMap(cfgMap map[string]interface{}, prefix string, driverKey string) (T, error) {
	var zero T

	// Get the driver
	driver, found := MapKeyStr(cfgMap, driverKey, true)
	if !found {
		return zero, fmt.Errorf("no driver found under key %v", driverKey)
	}
	if driver == "" {
		return zero, fmt.Errorf("empty driver found under key %v", driverKey)
	}

	factory, ok := MapKey(pr.Providers, driver, true)
	if !ok {
		return zero, ErrDriverNotFound
	}
	cfg, err := factory.ToConfig(cfgMap)
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

func (pr *ProvidersRegistry[T]) NewFromEnv(prefix string, driverKey string) (T, error) {
	var zero T
	env := LoadEnvPrefixMap(prefix)
	driver, found := MapKeyStr(env, driverKey, true)

	if !found {
		return zero, fmt.Errorf("no driver found under key %v", driverKey)
	}
	if driver == "" {
		return zero, fmt.Errorf("empty driver found under key %v", driverKey)
	}

	return pr.NewFromMap(env, prefix, driverKey)
}

func (pr *ProvidersRegistry[T]) NewFromJson(jsonData []byte, prefix string, driverKey string) (T, error) {
	var zero T

	container, err := gabs.ParseJSON(jsonData)
	if err != nil {
		return zero, err
	}

	if prefix != "" {
		container = container.Path(prefix)
	}
	if container == nil {
		return zero, fmt.Errorf("json key not found %v", prefix)
	}

	result := make(map[string]interface{})
	data := container.Bytes()
	err = json.Unmarshal(data, &result)
	if err != nil {
		return zero, err
	}

	return pr.NewFromMap(result, prefix, driverKey)
}

type Storage interface {
}

// Loads all the environment variables with the given prefix.
// This normalizes the env variable names to all uppercase
// without underscores. This means that POOL_ID == POOLID == PoolId == pool_id
func LoadEnvPrefixMap(prefix string) map[string]interface{} {
	all := os.Environ()
	uprefix := strings.ToUpper(prefix)

	rtn := make(map[string]interface{})
	for _, env := range all {
		// Replace all the _ characters
		noUnderscore := strings.ReplaceAll(env, "_", "")
		envu := strings.ToUpper(noUnderscore)

		if strings.HasPrefix(envu, uprefix) {
			key := envu[len(uprefix):]
			rtn[key] = os.Getenv(env)
		}
	}

	return rtn
}
