package datastore

import (
	"context"

	"github.com/appliedres/cloudy"
)

var UntypedJsonDataStoreFactoryProviders = cloudy.NewProviderRegistry[UntypedJsonDataStoreFactory]()

// Interface to create JSON datastores for a given type
type UntypedJsonDataStoreFactory interface {
	// CreateJsonDatastore creates a datastore for an individual typename. This is used
	// as a table name, collection name, etc. There can be a prefix provided to keep
	// any tables or collections prefixed
	CreateJsonDatastore(ctx context.Context, typename string, prefix string, idField string) UntypedJsonDataStore
}

func NewJsonDatastoreFromEnv[T any](ctx context.Context, env *cloudy.Environment, defaultDriver string) (UntypedJsonDataStoreFactory, error) {
	driver := env.Default("DRIVER", defaultDriver)
	ds, err := UntypedJsonDataStoreFactoryProviders.NewFromEnvWith(cloudy.DefaultEnvironment, driver)
	return ds, err
}

func CreateJsonDatastore[T any](ctx context.Context, name string, prefix string, idField string, env *cloudy.Environment) (JsonDataStore[T], error) {
	factory, err := UntypedJsonDataStoreFactoryProviders.NewFromEnv(env, "DRIVER")
	if err != nil {
		return nil, err
	}
	jds := factory.CreateJsonDatastore(ctx, name, prefix, idField)
	tds := NewTypedStore[T](jds)

	return tds, nil
}

type DatabaseConfig struct {
	Engine   string
	Password string
	User     string
	Host     string
	Database string
	Other    map[string]string
}
