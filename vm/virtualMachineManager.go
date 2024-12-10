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
type VirtualMachineManager interface {
	GetAll(ctx context.Context, filter string, attrs []string) (*[]models.VirtualMachine, error)

	// Retrieves a specific vm
	GetById(ctx context.Context, id string) (*models.VirtualMachine, error)

	// Create a new vm from a vm and returns it vm with any additional
	// fields populated
	Create(ctx context.Context, vm *models.VirtualMachine) (*models.VirtualMachine, error)

	// Update an existing vm from a vm and returns it with any additional
	// fields populated
	Update(ctx context.Context, vm *models.VirtualMachine) (*models.VirtualMachine, error)

	// Starts the vm with the provided id
	Start(ctx context.Context, id string) error

	// Stops the vm with the provided id
	Stop(ctx context.Context, id string) error

	// Deallocate the vm with the provided id
	Deallocate(ctx context.Context, id string) error

	// Deletes the vm with the provided id
	Delete(ctx context.Context, id string) error

	// Gets the vm size data with capabilities info filled in
	GetAllSizes(ctx context.Context) (map[string]*models.VirtualMachineSize, error)

	// Given a VM template, generates a ranked list of VM Sizes
	GetSizesForTemplate(ctx context.Context, template models.VirtualMachineTemplate) (map[string]*models.VirtualMachineSize, error)

	// Gets the vm size data with capabilities and usage info filled in
	GetSizesWithUsage(ctx context.Context) (map[string]*models.VirtualMachineSize, error)

	// Gets the vm family data with usage info filled in
	GetUsage(ctx context.Context) (map[string]models.VirtualMachineFamily, error)
}
