package cloudy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

var ErrDriverNotFound = errors.New("driver not found")
var ErrInvalidConfiguration = errors.New("invalid configuration object")
var ErrOperationNotImplemented = errors.New("operation not implemented")

type ProviderFactory[T any] interface {
	Create(cfg interface{}) (T, error)
	FromEnvMgr(em *EnvManager, prefix string) (interface{}, error)
}

type Provider[T any] struct {
	Factory      ProviderFactory[T]
	RequiredVars []EnvDefinition
}

type ProvidersRegistry[T any] struct {
	Providers map[string]Provider[T]
}

func NewProviderRegistry[T any]() *ProvidersRegistry[T] {
	return &ProvidersRegistry[T]{
		Providers: make(map[string]Provider[T]),
	}
}

func (pr *ProvidersRegistry[T]) Register(name string, factory ProviderFactory[T], requiredVars []EnvDefinition) {
	Info(context.Background(), "Provider.Register(): name=%s, requiredVars=%s", name, requiredVars)

	if _, exists := pr.Providers[name]; exists {
		log.Fatalf("Provider \"%s\" already registerd in this provider registry", name)
	}

	pr.Providers[name] = Provider[T]{
		Factory:      factory,
		RequiredVars: requiredVars, // change requiredVars to EnvDefinition type
	}

	if len(requiredVars) > 0 {
		for _, def := range requiredVars {
			Info(context.Background(), "\t\tprovider [%s]: adding provider var named \"%s\"", name, def.Name)
			DefaultEnvManager.RegisterDef(def) // TODO: provider var description
		}
	}
}

// get a list of vars required to create the provider
func (pr *ProvidersRegistry[T]) GetRequiredVars(em *EnvManager, driver string) ([]string, error) {
	Info(context.Background(), "Provider.GetRequiredVars(): driver=%s", driver)

	provider, ok := MapKey(pr.Providers, driver, true)
	if !ok {
		var keys []string
		for key := range pr.Providers {
			keys = append(keys, key)
		}
		Error(context.Background(), "GetRequiredVars: Driver \"%s\" Not found. Available drivers are: %v", driver, keys)
		return nil, ErrDriverNotFound
	}

	// TODO: pass keys or entire definition?
	keyMap := make(map[string]struct{})
	var keys []string

	for _, def := range provider.RequiredVars {
		if _, exists := keyMap[def.Key]; exists {
			log.Fatalf("GetRequiredVars: duplicate key found in required vars \"%s\"", def.Key)
		}
		keyMap[def.Key] = struct{}{}
		keys = append(keys, def.Key)
	}

	return keys, nil
}

func (pr *ProvidersRegistry[T]) New(name string, cfg interface{}) (T, error) {
	Info(context.Background(), "Provider.New(): name=%s, cfg=%s", name, cfg)

	var zero T
	provider, ok := MapKey(pr.Providers, name, true)
	if !ok {
		return zero, ErrDriverNotFound
	}

	return provider.Factory.Create(cfg)
}

func (pr *ProvidersRegistry[T]) NewFromEnvMgr(em *EnvManager, driverKey string) (T, error) {
	Info(context.Background(), "Provider.NewFromEnvMgr(): driverKey=%s", driverKey)

	var zero T

	driver := em.GetVar(driverKey+"_DRIVER", "DRIVER")
	provider, ok := MapKey(pr.Providers, driver, true)
	if !ok {
		var keys []string
		for key := range pr.Providers {
			keys = append(keys, key)
		}
		Error(context.Background(), "NewFromEnvMgr: Driver \"%s\" Not found. Available drivers are: %v", driver, keys)
		return zero, ErrDriverNotFound
	}

	// filteredEnvMgr := em.FilteredEnvManagerByKeyPrefix(driverKey)
	// cfg, err := provider.Factory.FromEnvMgr(filteredEnvMgr)

	cfg, err := provider.Factory.FromEnvMgr(em, driverKey)
	if err != nil {
		return zero, err
	}
	return provider.Factory.Create(cfg)
}

func (pr *ProvidersRegistry[T]) NewFromEnvMgrWith(em *EnvManager, driver string) (T, error) {
	Info(context.Background(), "Provider.NewFromEnvMgrWith(): driver=%s", driver)

	var zero T
	provider, ok := MapKey(pr.Providers, driver, true)
	if !ok {
		var keys []string
		for key := range pr.Providers {
			keys = append(keys, key)
		}
		Error(context.Background(), "NewFromEnvMgrWith: Driver \"%s\" Not found. Available drivers are: %v", driver, keys)
		return zero, ErrDriverNotFound
	}
	cfg, err := provider.Factory.FromEnvMgr(em, driver)
	if err != nil {
		return zero, err
	}
	return provider.Factory.Create(cfg)
}

func FromEnvMgr(prefix string, driverKey string) (string, map[string]interface{}, error) {
	Info(context.Background(), "Provider.FromEnvMgr(): prefix=%s, driver=%s", prefix, driverKey)

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
	Info(context.Background(), "Provider.LoadEnvPrefixMap(): prefix=%s", prefix)

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
