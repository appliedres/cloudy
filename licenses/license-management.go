package licenses

import (
	"context"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/models"
)

var LicenseProviders = cloudy.NewProviderRegistry[LicenseManager]()

type LicenseDescription struct {
	ID          string
	SKU         string
	Name        string
	Description string
	Assigned    int
	Total       int
}

type LicenseManager interface {
	//AssignLicense assign a license to a user. Maybe add the full user object here for things like gitlab
	AssignLicense(ctx context.Context, userdId string, licenseSkus ...string) error

	//RemoveLicense removes a license to a user. Maybe add the full user object here for things like gitlab
	RemoveLicense(ctx context.Context, userdId string, licenseSkus ...string) error

	// Get the licenses for a user
	GetUserAssigned(ctx context.Context, userdId string) ([]*LicenseDescription, error)

	//GetAssigned gets a list of all the users with licenses
	GetAssigned(ctx context.Context, licenseSku string) ([]*models.User, error)

	//ListLicenses List all the managed licenses
	ListLicenses(ctx context.Context) ([]*LicenseDescription, error)
}
