// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// VirtualMachineSize the details associated with the virtual machine family (most of the values are retrieved from the cloud)
//
// swagger:model VirtualMachineSize
type VirtualMachineSize struct {

	// flag for if this family of virtual machines can use accelerated networking
	AcceleratedNetworking bool `json:"acceleratedNetworking,omitempty"`

	// remaining number of this family of virtual machines that may be used
	Available int64 `json:"available,omitempty"`

	// the number of CPUs available to this virtual machine family
	CPU int64 `json:"cpu,omitempty"`

	// the vendor of the CPU of this virtual machine family
	CPUVendor string `json:"cpuVendor,omitempty"`

	// the name of the virtual machine family
	Description string `json:"description,omitempty"`

	// estimated cost per hour of this virtual machine family
	EstimatedCostPerHour float64 `json:"estimatedCostPerHour,omitempty"`

	// the vm size family
	Family *VirtualMachineFamily `json:"family,omitempty"`

	// the number of GPUs available to this virtual machine family
	Gpu int64 `json:"gpu,omitempty"`

	// the vendor of the GPU of this virtual machine family
	GpuVendor string `json:"gpuVendor,omitempty"`

	// the id of virtual machine family
	ID string `json:"id,omitempty"`

	// map of locations where this family is available
	Locations map[string]*VirtualMachineLocation `json:"locations,omitempty"`

	// maximum number of data disks that can be attached to this virtual machine family
	MaxDataDisks int64 `json:"maxDataDisks,omitempty"`

	// maximum amount of disk IO per second available to this virtual machine family
	MaxIops int64 `json:"maxIops,omitempty"`

	// maximum network bandwidth per second available to this virtual machine family
	MaxNetworkBandwidth int64 `json:"maxNetworkBandwidth,omitempty"`

	// maximum number of network interfaces that can be attached to this virtual machine family
	MaxNetworkInterfaces int64 `json:"maxNetworkInterfaces,omitempty"`

	// the name of the virtual machine family
	Name string `json:"name,omitempty"`

	// administrative notes concerning this virtual machine family (not from the cloud)
	Notes string `json:"notes,omitempty"`

	// flag for if this family of virtual machines can use premium IO
	PremiumIo bool `json:"premiumIo,omitempty"`

	// the amount of RAM in GB available to this virtual machine family
	RAM float64 `json:"ram,omitempty"`

	// flag for if this family of virtual machines may be used
	Restricted bool `json:"restricted,omitempty"`

	// tags
	Tags map[string]*string `json:"tags,omitempty"`
}

// Validate validates this virtual machine size
func (m *VirtualMachineSize) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFamily(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateLocations(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *VirtualMachineSize) validateFamily(formats strfmt.Registry) error {
	if swag.IsZero(m.Family) { // not required
		return nil
	}

	if m.Family != nil {
		if err := m.Family.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("family")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("family")
			}
			return err
		}
	}

	return nil
}

func (m *VirtualMachineSize) validateLocations(formats strfmt.Registry) error {
	if swag.IsZero(m.Locations) { // not required
		return nil
	}

	for k := range m.Locations {

		if err := validate.Required("locations"+"."+k, "body", m.Locations[k]); err != nil {
			return err
		}
		if val, ok := m.Locations[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("locations" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("locations" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this virtual machine size based on the context it is used
func (m *VirtualMachineSize) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateFamily(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateLocations(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *VirtualMachineSize) contextValidateFamily(ctx context.Context, formats strfmt.Registry) error {

	if m.Family != nil {
		if err := m.Family.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("family")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("family")
			}
			return err
		}
	}

	return nil
}

func (m *VirtualMachineSize) contextValidateLocations(ctx context.Context, formats strfmt.Registry) error {

	for k := range m.Locations {

		if val, ok := m.Locations[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *VirtualMachineSize) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VirtualMachineSize) UnmarshalBinary(b []byte) error {
	var res VirtualMachineSize
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
