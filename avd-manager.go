package cloudy

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

/*
AVD interface manager
*/
type AzureVirtualDesktopManager interface {
	PreRegister(ctx context.Context, vm *models.VirtualMachine) (hostPoolName, token *string, err error)

	PostRegister(ctx context.Context, vm *models.VirtualMachine, hpName string) (*models.VirtualMachine, error)

	Cleanup(ctx context.Context, vmID string) error
}
