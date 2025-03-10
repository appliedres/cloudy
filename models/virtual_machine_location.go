// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// VirtualMachineLocation virtual machine location
//
// swagger:model VirtualMachineLocation
type VirtualMachineLocation struct {

	// the cloud the virtual machine is located in
	Cloud string `json:"cloud,omitempty"`

	// the id of the virtual machine location
	ID string `json:"id,omitempty"`

	// the region of the cloud the virtual machine is located In
	Region string `json:"region,omitempty"`

	// the subscription associated with the virtual machine
	Subscription string `json:"subscription,omitempty"`
}

// Validate validates this virtual machine location
func (m *VirtualMachineLocation) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this virtual machine location based on context it is used
func (m *VirtualMachineLocation) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VirtualMachineLocation) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VirtualMachineLocation) UnmarshalBinary(b []byte) error {
	var res VirtualMachineLocation
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
