package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func SecretsTest(t *testing.T, ctx context.Context, provider SecretProvider) {
	// String
	err := provider.SaveSecret(ctx, "test", "super-secret")
	assert.Nil(t, err)

	secStr, err := provider.GetSecret(ctx, "test")
	assert.Nil(t, err)
	assert.Equal(t, secStr, "super-secret")

	err = provider.DeleteSecret(ctx, "test")
	assert.Nil(t, err)

	secStr2, err := provider.GetSecret(ctx, "test")
	assert.Nil(t, err)
	assert.Equal(t, secStr2, "")

	// Binary
	var sampleBinary = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	err = provider.SaveSecretBinary(ctx, "test-bin", sampleBinary)
	assert.Nil(t, err)

	secBin, err := provider.GetSecretBinary(ctx, "test-bin")
	assert.Nil(t, err)
	assert.Equal(t, secBin, sampleBinary)

	err = provider.DeleteSecret(ctx, "test-bin")
	assert.Nil(t, err)

	secBin2, err := provider.GetSecretBinary(ctx, "test-bin")
	assert.Nil(t, err)
	assert.Nil(t, secBin2)
}
