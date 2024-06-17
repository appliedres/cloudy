package providers

import (
	"context"
	"fmt"

	"github.com/appliedres/cloudy/keyvalue"
)

var KeyValueStoreProviders = make(map[string]keyvalue.KeyValueStoreFactory)

func NewKeyValueStore(ctx context.Context, prefix string, store *keyvalue.KeyValueAggregator, driver string) (keyvalue.KeyValueStore, error) {
	// Look up the Value
	factory := KeyValueStoreProviders[driver]
	if factory == nil {
		return nil, fmt.Errorf("%v keystore driver not found", driver)
	}
	// Create the configuration object
	config := factory.NewConfig()
	err := store.GetObj(prefix, config)
	if err != nil {
		return nil, err
	}
	// Create the Object
	return factory.New(ctx, config)
}

func init() {
	KeyValueStoreProviders["file"] = &keyvalue.FileKeyValueStoreFactory{}

}
