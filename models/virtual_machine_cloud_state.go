// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// VirtualMachineCloudState State of the virtual machine as retrieved from the cloud.
//
// swagger:model VirtualMachineCloudState
type VirtualMachineCloudState string

func NewVirtualMachineCloudState(value VirtualMachineCloudState) *VirtualMachineCloudState {
	return &value
}

// Pointer returns a pointer to a freshly-allocated VirtualMachineCloudState.
func (m VirtualMachineCloudState) Pointer() *VirtualMachineCloudState {
	return &m
}

const (

	// VirtualMachineCloudStateCreating captures enum value "creating"
	VirtualMachineCloudStateCreating VirtualMachineCloudState = "creating"

	// VirtualMachineCloudStateRunning captures enum value "running"
	VirtualMachineCloudStateRunning VirtualMachineCloudState = "running"

	// VirtualMachineCloudStateStopping captures enum value "stopping"
	VirtualMachineCloudStateStopping VirtualMachineCloudState = "stopping"

	// VirtualMachineCloudStateStopped captures enum value "stopped"
	VirtualMachineCloudStateStopped VirtualMachineCloudState = "stopped"

	// VirtualMachineCloudStateStarting captures enum value "starting"
	VirtualMachineCloudStateStarting VirtualMachineCloudState = "starting"

	// VirtualMachineCloudStateRestarting captures enum value "restarting"
	VirtualMachineCloudStateRestarting VirtualMachineCloudState = "restarting"

	// VirtualMachineCloudStateDeleting captures enum value "deleting"
	VirtualMachineCloudStateDeleting VirtualMachineCloudState = "deleting"

	// VirtualMachineCloudStateDeleted captures enum value "deleted"
	VirtualMachineCloudStateDeleted VirtualMachineCloudState = "deleted"

	// VirtualMachineCloudStateFailed captures enum value "failed"
	VirtualMachineCloudStateFailed VirtualMachineCloudState = "failed"

	// VirtualMachineCloudStateUnknown captures enum value "unknown"
	VirtualMachineCloudStateUnknown VirtualMachineCloudState = "unknown"
)

// for schema
var virtualMachineCloudStateEnum []interface{}

func init() {
	var res []VirtualMachineCloudState
	if err := json.Unmarshal([]byte(`["creating","running","stopping","stopped","starting","restarting","deleting","deleted","failed","unknown"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		virtualMachineCloudStateEnum = append(virtualMachineCloudStateEnum, v)
	}
}

func (m VirtualMachineCloudState) validateVirtualMachineCloudStateEnum(path, location string, value VirtualMachineCloudState) error {
	if err := validate.EnumCase(path, location, value, virtualMachineCloudStateEnum, true); err != nil {
		return err
	}
	return nil
}

// Validate validates this virtual machine cloud state
func (m VirtualMachineCloudState) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateVirtualMachineCloudStateEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validates this virtual machine cloud state based on context it is used
func (m VirtualMachineCloudState) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}