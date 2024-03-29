package cloudy

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

var InviteProviders = NewProviderRegistry[InviteManager]()

/*
Invite interface manager
*/
type InviteManager interface {
	CreateInvitation(ctx context.Context, user *models.User, emailInvite bool, inviteRedirectUrl string) error
}
