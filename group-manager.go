package cloudy

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

var GroupProviders = NewProviderRegistry[GroupManager]()

// Manages groups that users are part of.This can be seperate
// from the user manager or it can be the same.
type GroupManager interface {

	// List all the groups available
	ListGroups(ctx context.Context) ([]*models.Group, error)

	// Get all the groups for a single user
	GetUserGroups(ctx context.Context, uid string) ([]*models.Group, error)

	// Create a new Group
	NewGroup(ctx context.Context, grp *models.Group) (*models.Group, error)

	// Update a group. This is generally just the name of the group.
	UpdateGroup(ctx context.Context, grp *models.Group) (bool, error)

	// Get all the members of a group. This returns partial users only,
	// typically just the user id, name and email fields
	GetGroupMembers(ctx context.Context, grpId string) ([]*models.User, error)

	// Remove members from a group
	RemoveMembers(ctx context.Context, groupId string, userIds []string) error

	// Add member(s) to a group
	AddMembers(ctx context.Context, groupId string, userIds []string) error
}
