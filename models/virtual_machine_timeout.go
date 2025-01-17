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

// VirtualMachineTimeout types of virtual machine timeout categories
//
// swagger:model VirtualMachineTimeout
type VirtualMachineTimeout string

func NewVirtualMachineTimeout(value VirtualMachineTimeout) *VirtualMachineTimeout {
	return &value
}

// Pointer returns a pointer to a freshly-allocated VirtualMachineTimeout.
func (m VirtualMachineTimeout) Pointer() *VirtualMachineTimeout {
	return &m
}

const (

	// VirtualMachineTimeoutWorkinghours captures enum value "workinghours"
	VirtualMachineTimeoutWorkinghours VirtualMachineTimeout = "workinghours"

	// VirtualMachineTimeoutWeekdays captures enum value "weekdays"
	VirtualMachineTimeoutWeekdays VirtualMachineTimeout = "weekdays"

	// VirtualMachineTimeoutNever captures enum value "never"
	VirtualMachineTimeoutNever VirtualMachineTimeout = "never"
)

// for schema
var virtualMachineTimeoutEnum []interface{}

func init() {
	var res []VirtualMachineTimeout
	if err := json.Unmarshal([]byte(`["workinghours","weekdays","never"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		virtualMachineTimeoutEnum = append(virtualMachineTimeoutEnum, v)
	}
}

func (m VirtualMachineTimeout) validateVirtualMachineTimeoutEnum(path, location string, value VirtualMachineTimeout) error {
	if err := validate.EnumCase(path, location, value, virtualMachineTimeoutEnum, true); err != nil {
		return err
	}
	return nil
}

// Validate validates this virtual machine timeout
func (m VirtualMachineTimeout) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateVirtualMachineTimeoutEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validates this virtual machine timeout based on context it is used
func (m VirtualMachineTimeout) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
