package license

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

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
	AssignLicense(ctx context.Context, user *models.User, licenseSku string) error

	//RemoveLicense removes a license to a user. Maybe add the full user object here for things like gitlab
	RemoveLicense(ctx context.Context, user *models.User, licenseSku string) error

	//GetAssigned gets a list of all the users with licenses
	GetAssigned(ctx context.Context, licenseSku string) ([]*models.User, error)

	//ListLicenses List all the managed licenses
	ListLicenses(ctx context.Context) ([]*LicenseDescription, error)
}
