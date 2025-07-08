package vm

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

type VirtualMachineOptions struct {
}

/*
Vm interface manager
*/
type VirtualDesktopOrchestrator interface {
	GetAllVirtualMachines(ctx context.Context, attrs []string, includeState bool) (*[]models.VirtualMachine, error)

	// Retrieves a specific virtual machine
	GetVirtualMachine(ctx context.Context, id string, includeState bool) (*models.VirtualMachine, error)

	// Create a new virtual machine and returns it with any additional
	// fields populated
	CreateVirtualMachine(ctx context.Context, vm *models.VirtualMachine, apiKeySecret, agentBinaryURL *string) (*models.VirtualMachine, error)

	// Update an existing virtual machine and returns it with any additional
	// fields populated
	UpdateVirtualMachine(ctx context.Context, vm *models.VirtualMachine) (*models.VirtualMachine, error)

	// Starts the virtual machine
	StartVirtualMachine(ctx context.Context, vm *models.VirtualMachine) error

	// Stops (deallocates) the virtual machine
	StopVirtualMachine(ctx context.Context, vm *models.VirtualMachine) error

	// Deletes the virtual machine
	DeleteVirtualMachine(ctx context.Context, vm *models.VirtualMachine) error

	// Gets the virtual machine size data with capabilities info filled in
	GetAllVirtualMachineSizes(ctx context.Context) (map[string]*models.VirtualMachineSize, error)

	// Given a virtual machine template, generates a ranked list of virtual machine sizes
	GetVirtualMachineSizesForTemplate(ctx context.Context, template models.VirtualMachineTemplate) (
		matches map[string]*models.VirtualMachineSize,
		worse map[string]*models.VirtualMachineSize,
		better map[string]*models.VirtualMachineSize,
		err error)

	// Gets the virtual machine size data with capabilities and usage info filled in
	GetVirtualMachineSizesWithUsage(ctx context.Context) (map[string]*models.VirtualMachineSize, error)

	// Gets the virtual machine family data with usage info filled in
	GetVirtualMachineUsage(ctx context.Context) (map[string]models.VirtualMachineFamily, error)
}
