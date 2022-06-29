package testutil

import (
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
	"github.com/stretchr/testify/assert"
)

func TestUserManager(t *testing.T, umg cloudy.UserManager) {
	ctx := cloudy.StartContext()

	user := "test.user@skyborg.onmicrosoft.us"

	u, err := umg.GetUser(ctx, user)
	assert.Nil(t, err)
	assert.Nil(t, u, "Should not be there")

	u = &models.User{
		ID:                 user,
		UserName:           user,
		FirstName:          "test",
		LastName:           "user",
		DisplayName:        "Test User",
		Password:           "dont_ever_use_1234%^&*",
		MustChangePassword: true,
	}

	u2, err := umg.NewUser(ctx, u)
	assert.Nil(t, err)
	assert.NotNil(t, u2, "Should be there")

	u3, err := umg.GetUser(ctx, user)
	assert.Nil(t, err)
	assert.NotNil(t, u3, "Should be there")

	err = umg.Disable(ctx, user)
	assert.Nil(t, err)

	err = umg.Enable(ctx, user)
	assert.Nil(t, err)

	u3.JobTitle = "Automated Tester"
	umg.UpdateUser(ctx, u3)

	for {
		users, next, err := umg.ListUsers(ctx, nil, nil)
		assert.Nil(t, err)
		assert.True(t, len(users) > 0)

		if next == nil {
			break
		}
	}

	err = umg.DeleteUser(ctx, user)
	assert.Nil(t, err)

}
