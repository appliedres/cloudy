package keyvalue

import "context"

type ConfigFactory interface {
	CreateConfig(ctx context.Context, store SecureKeyValueStore, prefix string) (interface{}, error)
}

type FromKeyValuePrefix struct {
	v interface{}
}

func NewFromKeyValuePrefix(v interface{}) *FromKeyValuePrefix {
	return &FromKeyValuePrefix{v}
}

func (f *FromKeyValuePrefix) CreateConfig(ctx context.Context, store SecureKeyValueStore, prefix string) (interface{}, error) {
	all, err := store.GetAll()
	if err != nil {
		return nil, err
	}
	trimmed := TrimKeyPrefix(all, prefix)
	err = Unmarshal(trimmed, f.v)
	return f.v, err
}
