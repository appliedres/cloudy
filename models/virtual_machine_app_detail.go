// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// VirtualMachineAppDetail virtual machine app detail
//
// swagger:model VirtualMachineAppDetail
type VirtualMachineAppDetail struct {

	// the id of the app to be installed on the vm
	AppID string `json:"appId,omitempty"`

	// the id of the version of the app installed on the vm
	AppVersionID string `json:"appVersionId,omitempty"`
}

// Validate validates this virtual machine app detail
func (m *VirtualMachineAppDetail) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this virtual machine app detail based on context it is used
func (m *VirtualMachineAppDetail) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VirtualMachineAppDetail) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VirtualMachineAppDetail) UnmarshalBinary(b []byte) error {
	var res VirtualMachineAppDetail
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}