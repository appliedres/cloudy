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

// VirtualMachineSecurityTypes virtual machine security types
//
// swagger:model VirtualMachineSecurityTypes
type VirtualMachineSecurityTypes string

func NewVirtualMachineSecurityTypes(value VirtualMachineSecurityTypes) *VirtualMachineSecurityTypes {
	return &value
}

// Pointer returns a pointer to a freshly-allocated VirtualMachineSecurityTypes.
func (m VirtualMachineSecurityTypes) Pointer() *VirtualMachineSecurityTypes {
	return &m
}

const (

	// VirtualMachineSecurityTypesNone captures enum value "None"
	VirtualMachineSecurityTypesNone VirtualMachineSecurityTypes = "None"

	// VirtualMachineSecurityTypesConfidentialVM captures enum value "ConfidentialVM"
	VirtualMachineSecurityTypesConfidentialVM VirtualMachineSecurityTypes = "ConfidentialVM"

	// VirtualMachineSecurityTypesTrustedLaunch captures enum value "TrustedLaunch"
	VirtualMachineSecurityTypesTrustedLaunch VirtualMachineSecurityTypes = "TrustedLaunch"
)

// for schema
var virtualMachineSecurityTypesEnum []interface{}

func init() {
	var res []VirtualMachineSecurityTypes
	if err := json.Unmarshal([]byte(`["None","ConfidentialVM","TrustedLaunch"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		virtualMachineSecurityTypesEnum = append(virtualMachineSecurityTypesEnum, v)
	}
}

func (m VirtualMachineSecurityTypes) validateVirtualMachineSecurityTypesEnum(path, location string, value VirtualMachineSecurityTypes) error {
	if err := validate.EnumCase(path, location, value, virtualMachineSecurityTypesEnum, true); err != nil {
		return err
	}
	return nil
}

// Validate validates this virtual machine security types
func (m VirtualMachineSecurityTypes) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateVirtualMachineSecurityTypesEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validates this virtual machine security types based on context it is used
func (m VirtualMachineSecurityTypes) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}