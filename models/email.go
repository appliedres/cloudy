// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// Email email
//
// swagger:model Email
type Email struct {

	// authentication required
	AuthenticationRequired bool `json:"AuthenticationRequired,omitempty"`

	// from
	From string `json:"From,omitempty"`

	// host
	Host string `json:"Host,omitempty"`

	// password
	Password string `json:"Password,omitempty"`

	// port
	Port string `json:"Port,omitempty"`
}

// Validate validates this email
func (m *Email) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this email based on context it is used
func (m *Email) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Email) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Email) UnmarshalBinary(b []byte) error {
	var res Email
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}