package cloudy

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

var UserProviders = NewProviderRegistry[UserManager]()

/*
User interface manager
*/
type UserManager interface {
	ListUsers(ctx context.Context, page interface{}, filter interface{}) ([]*models.User, interface{}, error)

	// Retrieves a specific user.
	GetUser(ctx context.Context, uid string) (*models.User, error)

	// NewUser creates a new user with the given information and returns the new user with any additional
	// fields populated
	NewUser(ctx context.Context, newUser *models.User) (*models.User, error)

	UpdateUser(ctx context.Context, usr *models.User) error

	Enable(ctx context.Context, uid string) error

	Disable(ctx context.Context, uid string) error

	DeleteUser(ctx context.Context, uid string) error
}
