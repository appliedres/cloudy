package testutil

import (
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
	"github.com/stretchr/testify/assert"
)

func TestUserManager(t *testing.T, umg cloudy.UserManager) {
	ctx := cloudy.StartContext()

	domain := cloudy.DefaultEnvironment.Force("USER_DOMAIN")

	user := "test.user@" + domain

	u := &models.User{
		UID:         "Q049VGVzdEdyb3VwLENOPVVzZXJzLERDPWxkYXAsREM9c2NobmVpZGUsREM9ZGV2",
		Email:       user,
		Username:    "test-user",
		FirstName:   "test",
		LastName:    "user",
		DisplayName: "Test User",
	}

	ug, err := umg.GetUser(ctx, u.UID)
	assert.Nil(t, err)
	if ug != nil {
		err = umg.DeleteUser(ctx, u.UID)
		assert.NotNil(t, err, "Test user should not exists and test is unable to delete it")
	}

	u2, err := umg.NewUser(ctx, u)
	assert.Nil(t, err)
	assert.NotNil(t, u2, "Unable to create Test User ("+user+")")
	assert.Equal(t, u.FirstName, u2.FirstName)
	assert.Equal(t, u.LastName, u2.LastName)
	assert.Equal(t, u.DisplayName, u2.DisplayName)
	assert.Equal(t, u.Username, u2.Username)

	u3, err := umg.GetUser(ctx, user)
	assert.Nil(t, err)
	assert.NotNil(t, u3, "Unable to retrieve Test User ("+user+")")
	assert.Equal(t, u.FirstName, u3.FirstName)
	assert.Equal(t, u.LastName, u2.LastName)
	assert.Equal(t, u.DisplayName, u3.DisplayName)
	assert.Equal(t, u.Username, u3.Username)
	assert.Equal(t, u2.UID, u3.UID)

	err = umg.Disable(ctx, user)
	assert.Nil(t, err, "Unable to disable Test User ("+user+")")

	err = umg.Enable(ctx, user)
	assert.Nil(t, err, "Unable to enable Test User ("+user+")")

	u3.LastName = "Tester"
	err = umg.UpdateUser(ctx, u3)
	assert.Nil(t, err, "%v", err)

	u3g, err := umg.GetUser(ctx, user)
	assert.Nil(t, err)
	assert.Equal(t, u3.LastName, u3g.LastName, "Updated user ("+user+") failed to post LastName Update")

	// for {
	// 	users, next, err := umg.ListUsers(ctx, "", nil)
	// 	assert.Nil(t, err)
	// 	assert.True(t, len(users) > 0)

	// 	if next == nil {
	// 		break
	// 	}
	// }

	err = umg.DeleteUser(ctx, user)
	assert.Nil(t, err)

	external_user := "test@external.com"
	u4 := &models.User{
		UID:         "Q049VGVzdEdyb3VwLENOPVVzZXJzLERDPWxkYXAsREM9c2NobmVpZGUsREM9ZGV2",
		Email:       external_user,
		Username:    "externaltest-user",
		FirstName:   "externaltest",
		LastName:    "externaluser",
		DisplayName: "External Test User",
	}

	u5, err := umg.NewUser(ctx, u4)
	assert.NotNil(t, err, "%v", err)
	assert.Nil(t, u5, "Should be there")

	// // test ForceUserName
	// usernameToForce := fmt.Sprintf("test.Bubba.%x@%s", time.Now().UnixNano(), domain)

	// // if username for the ForiceUserName already exists, delete it.
	// ug, _ = umg.GetUser(ctx, usernameToForce)
	// if ug != nil {
	// 	_ = umg.DeleteUser(ctx, usernameToForce)
	// }

	// // test ForceUserName where user does not exist
	// xformed, exists, err := umg.ForceUserName(ctx, usernameToForce)
	// assert.Equal(t, usernameToForce, xformed)
	// assert.False(t, exists, "Forced user should not exist")
	// assert.Nil(t, err, "%v", err)

	// // create a user to test the ForceUserName where the user does exist
	// u = &models.User{
	// 	ID:                 usernameToForce,
	// 	UPN:                usernameToForce,
	// 	FirstName:          "externaltest",
	// 	LastName:           "externaluser",
	// 	DisplayName:        "External Test User",
	// 	Password:           "dont_ever_use_1234%^&*",
	// 	MustChangePassword: true,
	// }
	// // if the user does exist, we don't care
	// _, _ = umg.NewUser(ctx, u)

	// // test ForceUserName where user does exist
	// xformed, exists, err = umg.ForceUserName(ctx, usernameToForce)
	// assert.Equal(t, usernameToForce, xformed)
	// assert.True(t, exists, "Forced user exists")
	// assert.Nil(t, err, "%v", err)

	// // clean up the test user
	// err = umg.DeleteUser(ctx, usernameToForce)
	// assert.Nil(t, err)

}
