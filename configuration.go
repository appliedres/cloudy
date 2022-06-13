package cloudy

import (
	"context"
	"os"
)

var ConfigProviders = NewProviderRegistry[ConfigurationService]()

type ConfigurationService interface {
	GetString(ctx context.Context, name string) (string, error)
	GetInt(ctx context.Context, name string) (int, error)
	GetMap(ctx context.Context, prefix string) (map[string]interface{}, error)
}

//Autoconfig attempts to configure the system using the following approach
// Look for a "config.json" or "config.toml" or "config.env" file
func Autoconfig() error {
	configSource := os.Getenv("CLOUDY_CONFIG_SOURCE")
	if configSource == "" {
		configSource = "ENV, TOML, JSON, ENV_FILE"
	}

	configFilePrefix := os.Getenv("CLOUDY_CONFIG_PREFIX")
	if configFilePrefix == "" {
		configFilePrefix = "config"
	}

	// order

	return nil
}
