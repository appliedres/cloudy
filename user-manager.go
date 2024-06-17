package cloudy

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

var UserProviders = NewProviderRegistry[UserManager]()

type UserOptions struct {
	IncludeLastSignIn *bool
}

/*
User interface manager
*/
type UserManager interface {
	// ForceUserName takes a proposed user name, validates it and transforms it.
	// Then it checks to see if it is a real user
	// Returns: string - updated user name, bool - if the user exists, error - if an error is encountered
	ForceUserName(ctx context.Context, name string) (string, bool, error)

	ListUsers(ctx context.Context, page interface{}, filter interface{}) ([]*models.User, interface{}, error)

	// Retrieves a specific user.
	GetUser(ctx context.Context, uid string) (*models.User, error)

	// Retrieves a specific user.
	GetUserByEmail(ctx context.Context, email string, opts *UserOptions) (*models.User, error)

	// NewUser creates a new user with the given information and returns the new user with any additional
	// fields populated
	NewUser(ctx context.Context, newUser *models.User) (*models.User, error)

	UpdateUser(ctx context.Context, usr *models.User) error

	Enable(ctx context.Context, uid string) error

	Disable(ctx context.Context, uid string) error

	DeleteUser(ctx context.Context, uid string) error
}

type AvatarManager interface {
	GetProfilePicture(ctx context.Context, uid string) ([]byte, error)
	UploadProfilePicture(ctx context.Context, uid string, picture []byte) error
}
