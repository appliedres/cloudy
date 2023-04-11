package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
	"github.com/stretchr/testify/assert"
)

func TestUserManager(t *testing.T, umg cloudy.UserManager) {
	ctx := cloudy.StartContext()

	domain := cloudy.DefaultEnvironment.Force("USER_DOMAIN")

	user := "test.user@" + domain

	u := &models.User{
		ID:                 user,
		UPN:                user,
		FirstName:          "test",
		LastName:           "user",
		DisplayName:        "Test User",
		Password:           "dont_ever_use_1234%^&*",
		MustChangePassword: true,
	}

	ug, err := umg.GetUser(ctx, u.ID)
	assert.Nil(t, err)
	if ug != nil {
		err = umg.DeleteUser(ctx, u.ID)
		assert.NotNil(t, err, "Test user should not exists and test is unable to delete it")
	}

	u2, err := umg.NewUser(ctx, u)
	assert.Nil(t, err)
	assert.NotNil(t, u2, "Unable to create Test User ("+user+")")
	assert.Equal(t, u.FirstName, u2.FirstName)
	assert.Equal(t, u.LastName, u2.LastName)
	assert.Equal(t, u.DisplayName, u2.DisplayName)
	assert.Equal(t, "", u2.Password)
	assert.Equal(t, u.UPN, u2.UPN)

	u3, err := umg.GetUser(ctx, user)
	assert.Nil(t, err)
	assert.NotNil(t, u3, "Unable to retrieve Test User ("+user+")")
	assert.Equal(t, u.FirstName, u3.FirstName)
	assert.Equal(t, u.LastName, u2.LastName)
	assert.Equal(t, u.DisplayName, u3.DisplayName)
	assert.Equal(t, "", u3.Password)
	assert.Equal(t, u.UPN, u3.UPN)
	assert.Equal(t, u2.ID, u3.ID)

	err = umg.Disable(ctx, user)
	assert.Nil(t, err, "Unable to disable Test User ("+user+")")

	err = umg.Enable(ctx, user)
	assert.Nil(t, err, "Unable to enable Test User ("+user+")")

	u3.JobTitle = "Automated Tester"
	err = umg.UpdateUser(ctx, u3)
	assert.Nil(t, err, "%v", err)

	u3g, err := umg.GetUser(ctx, user)
	assert.Nil(t, err)
	assert.Equal(t, u3.JobTitle, u3g.JobTitle, "Updated user ("+user+") failed to post JobTitle Update")

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

	external_user := "test@notskyborg.com"
	u4 := &models.User{
		ID:                 external_user,
		UPN:                external_user,
		FirstName:          "externaltest",
		LastName:           "externaluser",
		DisplayName:        "External Test User",
		Password:           "dont_ever_use_1234%^&*",
		MustChangePassword: true,
	}

	u5, err := umg.NewUser(ctx, u4)
	assert.NotNil(t, err, "%v", err)
	assert.Nil(t, u5, "Should be there")

	// test ForceUserName
	usernameToForce := fmt.Sprintf("test.Bubba.%x@%s", time.Now().UnixNano(), domain)

	// if username for the ForiceUserName already exists, delete it.
	ug, _ = umg.GetUser(ctx, usernameToForce)
	if ug != nil {
		_ = umg.DeleteUser(ctx, usernameToForce)
	}

	// test ForceUserName where user does not exist
	xformed, exists, err := umg.ForceUserName(ctx, usernameToForce)
	assert.Equal(t, usernameToForce, xformed)
	assert.False(t, exists, "Forced user should not exist")
	assert.Nil(t, err, "%v", err)

	// create a user to test the ForceUserName where the user does exist
	u = &models.User{
		ID:                 usernameToForce,
		UPN:                usernameToForce,
		FirstName:          "externaltest",
		LastName:           "externaluser",
		DisplayName:        "External Test User",
		Password:           "dont_ever_use_1234%^&*",
		MustChangePassword: true,
	}
	// if the user does exist, we don't care
	_, _ = umg.NewUser(ctx, u)

	// test ForceUserName where user does exist
	xformed, exists, err = umg.ForceUserName(ctx, usernameToForce)
	assert.Equal(t, usernameToForce, xformed)
	assert.True(t, exists, "Forced user exists")
	assert.Nil(t, err, "%v", err)

	// clean up the test user
	err = umg.DeleteUser(ctx, usernameToForce)
	assert.Nil(t, err)

}
