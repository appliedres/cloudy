// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// VirtualMachineApp virtual machine app
//
// swagger:model VirtualMachineApp
type VirtualMachineApp struct {

	// groups who can view this template during virtual machine creation.
	AllowedGroupIds []string `json:"allowedGroupIds"`

	// users who can view this template during virtual machine creation.
	AllowedUserIds []string `json:"allowedUserIds"`

	// banner path
	BannerPath string `json:"bannerPath,omitempty"`

	// cpu requirement
	CPURequirement float64 `json:"cpuRequirement,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// disallowed
	Disallowed bool `json:"disallowed,omitempty"`

	// featured
	Featured bool `json:"featured,omitempty"`

	// gpu requirement
	GpuRequirement bool `json:"gpuRequirement,omitempty"`

	// icon path
	IconPath string `json:"iconPath,omitempty"`

	// id
	ID string `json:"id,omitempty"`

	// included virtual machine app ids
	IncludedVirtualMachineAppIds []string `json:"includedVirtualMachineAppIds"`

	// incompatible apps
	IncompatibleApps []string `json:"incompatibleApps"`

	// license type
	LicenseType string `json:"licenseType,omitempty"`

	// license Url
	LicenseURL string `json:"licenseUrl,omitempty"`

	// location
	Location *VirtualMachineLocation `json:"location,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// notes
	Notes string `json:"notes,omitempty"`

	// os type
	OsType []string `json:"osType"`

	// owner group Id
	OwnerGroupID string `json:"ownerGroupId,omitempty"`

	// owner user Id
	OwnerUserID string `json:"ownerUserId,omitempty"`

	// ram requirement
	RAMRequirement int64 `json:"ramRequirement,omitempty"`

	// salt path
	SaltPath string `json:"saltPath,omitempty"`

	// software Url
	SoftwareURL string `json:"softwareUrl,omitempty"`

	// tags
	Tags []string `json:"tags"`

	// version
	Version string `json:"version,omitempty"`
}

// Validate validates this virtual machine app
func (m *VirtualMachineApp) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLocation(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *VirtualMachineApp) validateLocation(formats strfmt.Registry) error {
	if swag.IsZero(m.Location) { // not required
		return nil
	}

	if m.Location != nil {
		if err := m.Location.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("location")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("location")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this virtual machine app based on the context it is used
func (m *VirtualMachineApp) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateLocation(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *VirtualMachineApp) contextValidateLocation(ctx context.Context, formats strfmt.Registry) error {

	if m.Location != nil {
		if err := m.Location.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("location")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("location")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *VirtualMachineApp) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VirtualMachineApp) UnmarshalBinary(b []byte) error {
	var res VirtualMachineApp
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
