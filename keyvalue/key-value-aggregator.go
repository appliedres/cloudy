package keyvalue

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/go-multierror"
)

// Compiler assertions
var _ WritableKeyValueStore = (*KeyValueAggregator)(nil)
var _ FilteredKeyValueStore = (*KeyValueAggregator)(nil)
var _ WritableSecureKeyValueStore = (*KeyValueAggregator)(nil)

func NewKeyValueAggregator(stores ...KeyValueStore) *KeyValueAggregator {
	return &KeyValueAggregator{Stores: stores}
}

func NewBasicKeyValueAggregator() *KeyValueAggregator {
	file := NewDefaultFileKeyValueStore()
	os := NewBasicOsEnvKeyValueStore()
	return NewKeyValueAggregator(file, os)
}

type KeyValueAggregator struct {
	Stores []KeyValueStore
}

func (kva *KeyValueAggregator) AddFirst(store KeyValueStore) {
	kva.Stores = append([]KeyValueStore{store}, kva.Stores...)
}

func (kva *KeyValueAggregator) Append(store KeyValueStore) {
	kva.Stores = append(kva.Stores, store)
}

func (kva *KeyValueAggregator) Simple() SimpleKV {
	return &SimpleKvStore{kv: kva}
}

func (kva *KeyValueAggregator) Delete(key string) error {
	var merr *multierror.Error
	for _, store := range kva.Stores {
		w, isWritable := store.(WritableKeyValueStore)
		if isWritable {
			err := w.Delete(key)
			merr = multierror.Append(merr, err)
		}
	}
	return merr.ErrorOrNil()
}

func (kva *KeyValueAggregator) GetAlt(keys ...string) (string, error) {
	var merr *multierror.Error
	for _, key := range keys {
		val, err := kva.Get(key)
		if err != nil {
			merr = multierror.Append(merr, err)
		}
		if val != "" {
			return val, merr.ErrorOrNil()
		}
	}
	return "", merr.ErrorOrNil()
}

func (kva *KeyValueAggregator) getInternal(key string) (string, KeyValueStore, error) {
	var merr *multierror.Error
	for _, kvstore := range kva.Stores {
		v, err := kvstore.Get(key)
		merr = multierror.Append(merr, err)
		if err != nil {
			continue
		}
		if v != "" {
			return v, kvstore, merr.ErrorOrNil()
		}
	}
	return "", nil, merr.ErrorOrNil()
}

// Iterate through each Key Value store and try to get a value. If there is an error then
// try the next one down the list. TODO: If a KV keeps giving an error then place it in a
// timeout for a bit of time
func (kva *KeyValueAggregator) Get(key string) (string, error) {
	val, _, err := kva.getInternal(key)
	return val, err
}

func (kva *KeyValueAggregator) GetAll() (map[string]string, error) {
	var merr *multierror.Error
	all := make(map[string]string)
	for i := len(kva.Stores) - 1; i >= 0; i-- {
		kvstore := kva.Stores[i]
		allmine, err := kvstore.GetAll()
		merr = multierror.Append(merr, err)
		if err != nil {
			continue
		}
		for k, v := range allmine {
			all[k] = v
		}
	}
	return all, merr.ErrorOrNil()
}

func (kva *KeyValueAggregator) GetWithPrefix2(prefix string, trim bool) ([]map[string]string, error) {
	var merr *multierror.Error
	var layers []map[string]string
	for _, store := range kva.Stores {
		m, err := GetWithPrefix(store, prefix)
		if err != nil {
			merr = multierror.Append(merr, err)
			continue
		}
		if trim {
			m = TrimKeyPrefix(m, prefix)
		}
		layers = append(layers, m)
	}
	return layers, merr.ErrorOrNil()
}

func GetWithPrefix(store KeyValueStore, prefix string) (map[string]string, error) {
	kvFiltered, is := store.(FilteredKeyValueStore)
	if is {
		return kvFiltered.GetWithPrefix(prefix)
	}

	temp, err := store.GetAll()
	if err != nil {
		return nil, err
	}
	mine := make(map[string]string)
	for k, v := range temp {
		if strings.HasPrefix(k, prefix) {
			mine[k] = v
		}
	}
	return mine, nil
}

func (kva *KeyValueAggregator) GetWithPrefix(prefix string) (map[string]string, error) {
	var merr *multierror.Error
	var err error
	all := make(map[string]string)
	for i := len(kva.Stores) - 1; i >= 0; i-- {
		kvstore := kva.Stores[i]
		var mine map[string]string
		kvFiltered, is := kvstore.(FilteredKeyValueStore)
		if is {
			mine, err = kvFiltered.GetWithPrefix(prefix)
			merr = multierror.Append(merr, err)
			if err != nil {
				continue
			}
		} else {
			temp, err := kvstore.GetAll()
			merr = multierror.Append(merr, err)
			if err != nil {
				continue
			}
			mine = make(map[string]string)
			for k, v := range temp {
				if strings.HasPrefix(k, prefix) {
					mine[k] = v
				}
			}
		}
		merr = multierror.Append(merr, err)
		if err != nil {
			continue
		}
		for k, v := range mine {
			all[k] = v
		}
	}
	return all, merr.ErrorOrNil()
}

// Set needs to be aware of the originial location for all keys. If not found then
// just use the top layer
func (kva *KeyValueAggregator) Set(key string, value string) error {
	found, store, err := kva.getInternal(key)
	// Log error and continue
	if err != nil {
		slog.Error("Error in KeyValue store: %v", err)
	}

	if found != "" && store != nil {
		writeable, is := store.(WritableKeyValueStore)
		if is {
			err = writeable.Set(key, value)
			if err == nil {
				return nil
			}
			slog.Error("Error in Writable KeyValue store: %v", err)
		}
	}

	for _, store := range kva.Stores {
		writeable, is := store.(WritableKeyValueStore)
		if is {
			err = writeable.Set(key, value)
			if err == nil {
				return nil
			}
			slog.Error("Error in Writable KeyValue store: %v", err)
		}
	}
	return errors.New("No writable store found")
}

func (kva *KeyValueAggregator) SetMany(items map[string]string) error {
	var merr *multierror.Error

	for k, v := range items {
		err := kva.Set(k, v)
		merr = multierror.Append(merr, err)
	}
	return merr.ErrorOrNil()
}

func (kva *KeyValueAggregator) GetSecure(key string) (strfmt.Password, error) {
	value, _, err := kva.getSecureInternal(key)
	return value, err
}

func (kva *KeyValueAggregator) SetSecure(key string, value strfmt.Password) error {
	value, store, err := kva.getSecureInternal(key)

	// Log error and continue
	slog.Error("Error in KeyValue store: %v", err)

	if value != "" && store != nil {
		writeable, is := store.(WritableSecureKeyValueStore)
		if is {
			err = writeable.SetSecure(key, value)
			if err == nil {
				return nil
			}
			slog.Error("Error in Writable secure KeyValue store: %v", err)
		}
	}

	for _, store := range kva.Stores {
		writeable, is := store.(WritableSecureKeyValueStore)
		if is {
			err = writeable.SetSecure(key, value)
			if err == nil {
				return nil
			}
			slog.Error("Error in Writable secure KeyValue store: %v", err)
		}
	}
	return errors.New("No writable secure store found")
}

func (kva *KeyValueAggregator) getSecureInternal(key string) (strfmt.Password, KeyValueStore, error) {
	var merr *multierror.Error
	for _, store := range kva.Stores {
		secure, is := store.(SecureKeyValueStore)
		if is {
			v, err := secure.GetSecure(key)
			merr = multierror.Append(merr, err)
			if err != nil {
				continue
			}
			if v != "" {
				return v, store, merr.ErrorOrNil()
			}
		}
	}
	return "", nil, merr.ErrorOrNil()
}

func (kva *KeyValueAggregator) SetManyWithPrefix(prefix string, data map[string]interface{}) error {
	var merr *multierror.Error
	for k, v := range data {
		var key string
		if prefix == "" {
			key = fmt.Sprintf("%v", k)
		} else {
			key = fmt.Sprintf("%v.%v", prefix, k)
		}

		val := fmt.Sprintf("%v", v)
		err := kva.Set(key, val)
		merr = multierror.Append(merr, err)
	}
	return merr.ErrorOrNil()
}

func (kva *KeyValueAggregator) SetObj(name string, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	data2, _ := json.MarshalIndent(obj, "", "  ")
	fmt.Println(string(data2))

	c, err := gabs.ParseJSON(data)
	if err != nil {
		return err
	}
	flat, err := c.Flatten()
	if err != nil {
		return err
	}

	// Trim nils.
	for k, v := range flat {
		if v == nil {
			delete(flat, k)
		}
	}
	// common.PrintMap(flat)

	return kva.SetManyWithPrefix(name, flat)
}

func (kva *KeyValueAggregator) GetObjWithAlt(name string, v interface{}, altKeys map[string]string) error {
	allKeys, err := kva.GetAll()
	if allKeys == nil {
		return err
	}

	trimmed := TrimKeyPrefix(allKeys, name)
	AlternateKeys(trimmed, altKeys)
	return Unmarshal(trimmed, v)
}

func (kva *KeyValueAggregator) GetObj(prefix string, v interface{}) error {
	// Get ALL with this name
	var err error
	allMaps, err := kva.GetWithPrefix2(prefix, true)
	if allMaps == nil {
		return err
	}

	if len(allMaps) == 0 {
		return errors.New("No Data")
	}

	c := gabs.New()
	for i, m := range allMaps {
		for k, v := range m {
			child := c.Path(k)
			if child == nil {
				_, err := c.SetP(v, k)
				if err != nil {
					fmt.Printf("Error setting %v, %v\n", k, v)
				}
			}
		}

		fmt.Printf("AFTER %v\n", i)
		fmt.Println(c.StringIndent("", "  "))
	}

	fmt.Println(c.StringIndent("", "  "))
	return json.Unmarshal(c.Bytes(), v)
}

func TrimKeyPrefix(in map[string]string, prefix string) map[string]string {

	if prefix == "" {
		return in
	}

	p := prefix + "."
	m2 := make(map[string]string)
	for k, v := range in {
		if strings.HasPrefix(k, p) {
			newKey := strings.TrimPrefix(k, p)
			m2[newKey] = v
		}
	}
	return m2
}

// Searches the provided map for any alternate Keys.
func AlternateKeys(m map[string]string, alternateKeys map[string]string) {
	for alt, dst := range alternateKeys {
		AlternateKey(m, alt, dst)
	}
}

func AlternateKey(m map[string]string, alt string, dst string) {
	_, found := m[dst]
	if found {
		return
	}

	val, found := m[alt]
	if found {
		m[dst] = val
	}
}

func NormalizeKey(key string) string {
	// Lowercase
	// hypens to dots
	// underscores to dots
	rtn := strings.ToLower(key)
	rtn = strings.ReplaceAll(rtn, "-", ".")
	rtn = strings.ReplaceAll(rtn, "_", ".")
	rtn = strings.ReplaceAll(rtn, " ", ".")

	return rtn
}

func Unmarshal(data map[string]string, v any) error {
	c := gabs.New()
	for k, v := range data {
		_, err := c.SetP(v, k)
		if err != nil {
			fmt.Printf("Error with path %v, %v", k, err)
			// return err
		}
	}

	result := c.Bytes()
	return json.Unmarshal(result, v)
}

type SimpleKV interface {
	Get(key string, altKeys ...string) string
	GetP(key string, altKeys ...string) *string
	Default(defaultValue string, key string, altKeys ...string) string
	DefaultP(defaultValue string, key string, altKeys ...string) *string
	Error() error
}

var _ SimpleKV = (*SimpleKvStore)(nil)

type SimpleKvStore struct {
	kv   KeyValueStore
	merr *multierror.Error
}

func (s *SimpleKvStore) Error() error {
	err := s.merr.ErrorOrNil()
	s.merr = nil
	return err
}

func (s *SimpleKvStore) GetP(key string, altKeys ...string) *string {
	found := s.Get(key, altKeys...)
	if found == "" {
		return nil
	}
	return &found
}

func (s *SimpleKvStore) Get(key string, altKeys ...string) string {
	all := append([]string{key}, altKeys...)
	for _, k := range all {
		found, err := s.kv.Get(k)
		if err != nil {
			s.merr = multierror.Append(err)
			continue
		}
		if found != "" {
			return found
		}
	}
	return ""
}

func (s *SimpleKvStore) Default(defaultValue string, key string, altKeys ...string) string {
	found := s.Get(key, altKeys...)
	if found == "" {
		return defaultValue
	}
	return found
}

func (s *SimpleKvStore) DefaultP(defaultValue string, key string, altKeys ...string) *string {
	found := s.Default(defaultValue, key, altKeys...)
	if found == "" {
		return nil
	}
	return &found
}
