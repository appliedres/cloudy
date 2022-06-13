package secrets

import (
	"context"
	"encoding/json"

	"github.com/appliedres/cloudy"
)

var SecretProviders = cloudy.NewProviderRegistry[SecretProvider]()

type SecretProvider interface {
	SaveSecretBinary(ctx context.Context, key string, secret []byte) error
	GetSecretBinary(ctx context.Context, key string) ([]byte, error)

	GetSecret(ctx context.Context, key string) (string, error)
	SaveSecret(ctx context.Context, key string, data string) error

	DeleteSecret(ctx context.Context, key string) error
}

func SaveSecret[T any](ctx context.Context, driver SecretProvider, key string, value *T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return driver.SaveSecret(ctx, key, string(data))
}

func GetSecret[T any](ctx context.Context, driver SecretProvider, key string) (*T, error) {
	data, err := driver.GetSecret(ctx, key)
	if err != nil {
		return nil, err
	}
	var item T
	err = json.Unmarshal([]byte(data), &item)
	return &item, err
}
