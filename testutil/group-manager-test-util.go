package testutil

import (
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
	"github.com/stretchr/testify/assert"
)

func TestGroupManager(t *testing.T, gm cloudy.GroupManager, memberId string) {
	ctx := cloudy.StartContext()

	grps, err := gm.ListGroups(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, grps)

	// Now check to see if we have a "UNIT_TEST" group
	var testG string
	for _, g := range grps {
		if g.Name == "UNIT_TEST" {
			testG = g.ID
			break
		}
	}

	if testG == "" {
		newG, err := gm.NewGroup(ctx, &models.Group{
			Name: "UNIT_TEST",
		})
		assert.Nil(t, err)
		assert.NotNil(t, newG)
		if newG != nil {
			testG = newG.ID
		}
	}

	if testG == "" {
		t.FailNow()
	}

	// Add some users
	err = gm.AddMembers(ctx, testG, []string{memberId})
	assert.Nil(t, err)

	// Get the members
	people, err := gm.GetGroupMembers(ctx, testG)
	assert.Nil(t, err)
	assert.NotNil(t, people)
	var found *models.User
	for _, u := range people {
		if u.ID == memberId {
			found = u
			break
		}
	}
	assert.NotNil(t, found)

	// Makes sure

	usergroups, err := gm.GetUserGroups(ctx, memberId)
	assert.Nil(t, err)
	assert.True(t, len(usergroups) > 0)

	// Remove
	err = gm.RemoveMembers(ctx, testG, []string{memberId})
	assert.Nil(t, err)

	people2, err := gm.GetGroupMembers(ctx, testG)
	assert.Nil(t, err)
	assert.NotNil(t, people2)
	var found2 *models.User
	for _, u := range people2 {
		if u.UserName == memberId {
			found2 = u
			break
		}
	}
	assert.Nil(t, found2)

	err = gm.DeleteGroup(ctx, testG)
	assert.Nil(t, err)

	grpDeleted, err := gm.GetGroup(ctx, testG)
	assert.Nil(t, err)
	assert.Nil(t, grpDeleted)

}
