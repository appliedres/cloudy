package cloudy

import (
	"context"
)

var AVDProviders = NewProviderRegistry[AVDManager]()

/*
AVD interface manager
*/
type AVDManager interface {
	// first 4 methods needed to add a VM to AVD and assign to a user
	// find a host pool in the resource group that the user does not have a direct VM in, returns host pool name
	FindFirstAvailableHostPool(ctx context.Context, rg string, upn string) (*string, error)

	// get the registration key using host pool name, returns reg token
	RetrieveRegistrationToken(ctx context.Context, rg string, hpname string) (*string, error)

	// assign a session host to a user (currently must use user object id)
	AssignSessionHost(ctx context.Context, rg string, hpname string, sessionhost string, userobjectid string) error

	// needs to be called twice to assign to user to following roles at resouce group level, need object id of the role
	// Desktop Virtualization User
	// Virtual Machine User Login
	AssignRoleToUser(ctx context.Context, rg string, roleid string, upn string) error

	// utility methods
	// remove session host from AVD, does not delete VM
	DeleteSessionHost(ctx context.Context, rg string, hpname string, sessionhost string) error

	// Delete a user from a session host, user would have to be re assigned to use VM
	DeleteUserSession(ctx context.Context, rg string, hpname string, sessionHost string, upn string) error

	// Disconnect a user from a session host, user still assigned to VM
	DisconnecteUserSession(ctx context.Context, rg string, hpname string, sessionHost string, upn string) error
}
