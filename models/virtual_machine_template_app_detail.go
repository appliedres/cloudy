// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// VirtualMachineTemplateAppDetail virtual machine template app detail
//
// swagger:model VirtualMachineTemplateAppDetail
type VirtualMachineTemplateAppDetail struct {

	// the id of the app to be installed on the vm
	AppID string `json:"appId,omitempty"`

	// Optional - the id of the version of the app to be installed on the vm
	AppVersionID string `json:"appVersionId,omitempty"`
}

// Validate validates this virtual machine template app detail
func (m *VirtualMachineTemplateAppDetail) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this virtual machine template app detail based on context it is used
func (m *VirtualMachineTemplateAppDetail) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VirtualMachineTemplateAppDetail) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VirtualMachineTemplateAppDetail) UnmarshalBinary(b []byte) error {
	var res VirtualMachineTemplateAppDetail
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}