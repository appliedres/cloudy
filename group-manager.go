package cloudy

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

var GroupProviders = NewProviderRegistry[Users]()

type Groups interface {
	ListGroups(ctx context.Context, uid string) ([]*models.Group, error)

	GetUserGroups(ctx context.Context, page interface{}, filter interface{}) ([]*models.Group, interface{}, error)

	NewGroup(ctx context.Context, grp *models.Group) (*models.Group, error)

	UpdateGroup(ctx context.Context, grp *models.Group) (bool, error)

	GetGroupMembers(ctx context.Context, grpId string) ([]string, error)
}
