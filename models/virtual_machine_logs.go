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

// VirtualMachineLogs virtual machine logs
//
// swagger:model VirtualMachineLogs
type VirtualMachineLogs struct {

	// id of the log (timestamp+vmid)
	ID string `json:"id,omitempty"`

	// the log text
	Log string `json:"log,omitempty"`

	// the time the log was recorded
	// Format: datetime
	Timestamp strfmt.DateTime `json:"timestamp,omitempty"`

	// id of the vm associated with the log
	VMID string `json:"vmId,omitempty"`
}

// Validate validates this virtual machine logs
func (m *VirtualMachineLogs) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateTimestamp(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *VirtualMachineLogs) validateTimestamp(formats strfmt.Registry) error {
	if swag.IsZero(m.Timestamp) { // not required
		return nil
	}

	if err := validate.FormatOf("timestamp", "body", "datetime", m.Timestamp.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this virtual machine logs based on context it is used
func (m *VirtualMachineLogs) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VirtualMachineLogs) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VirtualMachineLogs) UnmarshalBinary(b []byte) error {
	var res VirtualMachineLogs
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}