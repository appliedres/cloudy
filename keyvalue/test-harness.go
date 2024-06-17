package keyvalue

import (
	"fmt"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
)

var TestStoreNormalForms map[string]string = map[string]string{
	"ENV_VALUE":                   "env.value",
	"env.value":                   "env.value",
	"env-value":                   "env.value",
	"Env Value":                   "env.value",
	"VMC_AZ_TENTANTID":            "vmc.az.tenantid",
	"StrangeKey.Value-With_Mixed": "strangekey.value.with.mixed",
}

// Test a KeyValue store implementation
func TestKVStore(t *testing.T, store KeyValueStore, storeValues map[string]string) {
	// Get Individual
	for k, v := range storeValues {
		found, err := store.Get(k)
		assert.NoError(t, err)
		assert.NotZero(t, found)
		assert.Equal(t, v, found)
	}

	// Bad Get
	notfound, err := store.Get("Really_Bad.Stupid-Key")
	assert.NoError(t, err)
	assert.Zero(t, notfound)

	// Get All
	all, err := store.GetAll()
	assert.NoError(t, err)
	assert.NotZero(t, all)
	AssertMapContainsNormalize(t, all, storeValues)
}

func TestWritableKVStore(t *testing.T, store WritableKeyValueStore, storeValues map[string]string) {
	// Test set one
	for k, v := range storeValues {
		err := store.Set(k, v)
		assert.NoError(t, err)

		found, err := store.Get(k)
		assert.NoError(t, err)
		assert.NotZero(t, found)
		assert.Equal(t, v, found)

		normKey := NormalizeKey(k)
		foundNorm, err := store.Get(normKey)
		assert.NoError(t, err)
		assert.NotZero(t, foundNorm)
		assert.Equal(t, v, foundNorm)

		err = store.Delete(k)
		assert.NoError(t, err)

		notfound, err := store.Get(k)
		assert.NoError(t, err)
		assert.Zero(t, notfound)
	}

	// Test set Many
	err := store.SetMany(storeValues)
	assert.NoError(t, err)

	all, err := store.GetAll()
	assert.NoError(t, err)

	AssertMapContainsNormalize(t, all, storeValues)

}

func TestFilteredKVStore(t *testing.T, store FilteredKeyValueStore) {

}

func TestSecureKVStore(t *testing.T, store SecureKeyValueStore) {

}

func TestWritableSecureKVStore(t *testing.T, store WritableSecureKeyValueStore) {
	value := strfmt.Password("1234QWER!@#$")
	key := "secure-key"
	err := store.SetSecure(key, value)
	assert.NoError(t, err)

	secureValue, err := store.GetSecure(key)
	assert.NoError(t, err)
	assert.NotZero(t, secureValue)
	assert.Equal(t, value, secureValue)

	err = store.Delete(key)
	assert.NoError(t, err)

	notfound, err := store.GetSecure(key)
	assert.NoError(t, err)
	assert.Zero(t, notfound)
}

func AssertMapEqual[V comparable](t *testing.T, m1 map[string]V, m2 map[string]V) {
	AssertMapContainsNormalize(t, m1, m2)
	AssertMapContainsNormalize(t, m2, m1)
}

func AssertMapContainsNormalize[V comparable](t *testing.T, all map[string]V, contains map[string]V) {
	for k, v := range contains {
		nkey := NormalizeKey(k)
		v2, ok := all[nkey]
		assert.True(t, ok)
		assert.Equal(t, v, v2)
		fmt.Printf("Comparing [%v]-->[%v] [%v]-->[%v]\n", k, v, nkey, v2)
	}
}
