package testutil

import (
	"context"
	"testing"

	"github.com/appliedres/cloudy/licenses"
	"github.com/stretchr/testify/assert"
)

func TestLicenseManager(t *testing.T, ctx context.Context, lm licenses.LicenseManager, userId string, sku string) {

	all, err := lm.ListLicenses(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, all)

	assigned, err := lm.GetUserAssigned(ctx, userId)
	assert.Nil(t, err)

	// Check to see
	found := false
	for _, a := range assigned {
		if a.SKU == sku {
			found = true
			break
		}
	}
	foundOriginal := found

	if !found {
		err = lm.AssignLicense(ctx, userId, sku)
		assert.Nil(t, err)
	}

	// Now check again
	assigned, err = lm.GetUserAssigned(ctx, userId)
	assert.Nil(t, err)
	found = false
	for _, a := range assigned {
		if a.SKU == sku {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Make sure we show some assigned
	allFound, err := lm.GetAssigned(ctx, sku)
	assert.Nil(t, err)

	assert.NotNil(t, allFound)

	// Now remove
	err = lm.RemoveLicense(ctx, userId, sku)
	assert.Nil(t, err)

	// CHeck again
	assigned, err = lm.GetUserAssigned(ctx, userId)
	assert.Nil(t, err)
	found = false
	for _, a := range assigned {
		if a.SKU == sku {
			found = true
			break
		}
	}
	assert.False(t, found)

	// If it was removed originally then set it back
	if foundOriginal {
		err = lm.AssignLicense(ctx, userId, sku)
		assert.Nil(t, err)
	}

}
