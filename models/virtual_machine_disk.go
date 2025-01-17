// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// VirtualMachineDisk virtual machine disk
//
// swagger:model VirtualMachineDisk
type VirtualMachineDisk struct {

	// estimated cost per hour for the disk (based on size)
	EstimatedCostPerHour float64 `json:"estimatedCostPerHour,omitempty"`

	// full path id of the disk
	ID string `json:"id,omitempty"`

	// name of the disk
	Name string `json:"name,omitempty"`

	// flag is true for if this disk is an operating system disk
	OsDisk bool `json:"osDisk,omitempty"`

	// flag is true if the virtual machine disk has premium IO enabled
	PremiumIo bool `json:"premiumIo,omitempty"`

	// disk size in GB
	Size int64 `json:"size,omitempty"`
}

// Validate validates this virtual machine disk
func (m *VirtualMachineDisk) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this virtual machine disk based on context it is used
func (m *VirtualMachineDisk) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VirtualMachineDisk) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VirtualMachineDisk) UnmarshalBinary(b []byte) error {
	var res VirtualMachineDisk
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
