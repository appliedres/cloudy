package cloudy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

func init() {
	EnvironmentProviders.Register("osenv", &SystemEnvironmentVariablesFactory{})
	EnvironmentProviders.Register("test", &TestEnvFileFactory{})
}

var ErrKeyNotFound = errors.New("key not found")

type CachedEnvironment struct {
	c      *cache.Cache
	Source EnvironmentService
}

func NewCachedEnvironment(source EnvironmentService) *CachedEnvironment {
	return &CachedEnvironment{
		c:      cache.New(5*time.Minute, 10*time.Minute),
		Source: source,
	}
}

func (ce *CachedEnvironment) Get(name string) (string, error) {
	value, found := ce.c.Get(name)
	if !found {
		sValue, sErr := ce.Source.Get(name)
		if sErr == nil {
			ce.c.Set(name, sValue, 0)
			return sValue, nil
		}
		return "", ErrKeyNotFound
	}
	return value.(string), nil
}

// System Environment Variables Services
type SystemEnvironmentVariables struct{}

func NewOsEnvironmentService() *SystemEnvironmentVariables {
	return &SystemEnvironmentVariables{}
}

func (ce *SystemEnvironmentVariables) Get(name string) (string, error) {
	value, found := os.LookupEnv(name)
	if !found {
		return "", ErrKeyNotFound
	}
	return value, nil
}

type SystemEnvironmentVariablesFactory struct{}

func (f *SystemEnvironmentVariablesFactory) Create(cfg interface{}) (EnvironmentService, error) {
	return &SystemEnvironmentVariables{}, nil
}
func (f *SystemEnvironmentVariablesFactory) FromEnv(env *Environment) (interface{}, error) {
	return nil, nil
}

/*
TieredEnvironment provides an EnvironmentService that looks for
configuration values in a chain of environments, starting at
the first.

For Example, consider the following

	env := NewTieredEnvironment(
		NewSystemEnvironmentVariables(),
		NewCachedEnvironment(NewKeyVaultEnvironmentService(myazconfig))
	)

	val, err := env.Get("A-Value")

In this example. The cached tier is used first
*/
type TieredEnvironment struct {
	sources []EnvironmentService
}

func NewTieredEnvironment(sources ...EnvironmentService) *TieredEnvironment {
	return &TieredEnvironment{
		sources: sources,
	}
}

func (te *TieredEnvironment) Get(name string) (string, error) {
	Info(context.Background(), "TieredEnvironment Get: %s", name)

	for _, env := range te.sources {
		val, err := env.Get(name)
		if err == nil {
			return val, nil
		}
	}
	return "", ErrKeyNotFound
}

func (te *TieredEnvironment) Force(name string) (string, error) {
	Info(context.Background(), "TieredEnvironment Force: %s", name)

	for _, env := range te.sources {
		val, err := env.Get(name)
		if err == nil {
			return val, nil
		}
	}

	log.Fatalf("TieredEnvironment: Force Required Variable not found, %s", name)

	return "", ErrKeyNotFound

}

type MapEnvironment struct {
	data map[string]string
}

func NewMapEnvironment() *MapEnvironment {
	return &MapEnvironment{
		data: make(map[string]string),
	}
}

func (te *MapEnvironment) Get(name string) (string, error) {
	val, found := te.data[name]

	if !found {
		return "", ErrKeyNotFound
	}

	return val, nil
}

func (te *MapEnvironment) Set(name string, v string) {
	te.data[name] = v
}

func LoadEnvironmentServiceFromString(data string) (*MapEnvironment, error) {
	env := NewMapEnvironment()

	lines := strings.Split(data, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Comment
			continue
		}
		if len(trimmed) == 0 {
			continue
		}

		index := strings.Index(trimmed, "=")
		if index > 0 {
			k := trimmed[0:index]
			v := trimmed[index+1:]

			name := NormalizeEnvName(k)
			env.Set(name, v)
		}

	}
	return env, nil
}

func LoadEnvironmentService(file string) (*MapEnvironment, error) {
	fmt.Println(file)

	data, err := os.ReadFile(file)
	if err != nil {
		return NewMapEnvironment(), err
	}

	return LoadEnvironmentServiceFromString(string(data))
}

func NewCIEnvironmentService() *MapEnvironment {
	envCI := os.Getenv("ARKLOUD_ENV_CI")
	if envCI != "" {
		mp, err := LoadEnvironmentServiceFromString(envCI)
		if err != nil {
			fmt.Printf("Unable to load environment CI data... %v\n", err)
		} else {
			fmt.Printf("Loaded  %v from ARKLOUD_ENV_CI\n", len(mp.data))
			return mp
		}
	}
	return nil
}

func NewTestFileEnvironmentService() *MapEnvironment {
	// Loads the environment from a Environment variable (generally set in the CI)
	mp := NewCIEnvironmentService()
	if mp != nil {
		return mp
	}

	// Now check if there is a file
	envFilePath := os.Getenv("ARKLOUD_ENVFILE")
	if envFilePath == "" {
		currentDir, _ := os.Getwd()
		envFilePath = filepath.Join(currentDir, "test.env")
	}

	mp, err := LoadEnvironmentService(envFilePath)
	if err != nil {
		fmt.Printf("Unable to load %s environment file... this is ok\n", envFilePath)
	}

	return mp
}

type TestEnvFileFactory struct{}

func (f *TestEnvFileFactory) Create(cfg interface{}) (EnvironmentService, error) {
	return NewTestFileEnvironmentService(), nil
}

func (f *TestEnvFileFactory) FromEnv(env *Environment) (interface{}, error) {
	return nil, nil
}
