package cloudy

import (
	"context"
	"os"
	"strconv"
)

const ENV_PROVIDER = "env"

func init() {
	ConfigProviders.Register(ENV_PROVIDER, &EnvConfigFactory{})
}

type EnvConfigFactory struct{}

func (envFact *EnvConfigFactory) Create(cfg interface{}) (ConfigurationService, error) {
	cfgObg := cfg.(EnvConfigurationService)
	return &cfgObg, nil
}
func (envFact *EnvConfigFactory) ToConfig(config map[string]interface{}) (interface{}, error) {
	p, _ := MapKeyStr(config, "prefix", true)
	return &EnvConfigurationService{
		Prefix: p,
	}, nil
}

type EnvConfigurationService struct {
	Prefix string
}

func (svc *EnvConfigurationService) GetString(ctx context.Context, name string) (string, error) {
	fullName := svc.Prefix + name
	return os.Getenv(fullName), nil
}

func (svc *EnvConfigurationService) GetInt(ctx context.Context, name string) (int, error) {
	fullName := svc.Prefix + name
	val := os.Getenv(fullName)

	if val == "" {
		return 0, nil
	}

	return strconv.Atoi(val)
}

func (svc *EnvConfigurationService) GetMap(ctx context.Context, prefix string) (map[string]interface{}, error) {
	return LoadEnvPrefixMap(svc.Prefix), nil
}
