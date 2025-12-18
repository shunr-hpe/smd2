// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package component

import (
	"context"
	"encoding/json"

	"github.com/openchami/fabrica/pkg/resource"
)

// Component represents a Component resource
type Component struct {
	resource.Resource
	Spec   ComponentSpec   `json:"spec" validate:"required"`
	Status ComponentStatus `json:"status,omitempty"`
}

// ComponentSpec defines the desired state of Component
type ComponentSpec struct {
	Description string `json:"description,omitempty" validate:"max=200"`
	// Add your spec fields here
	ID                  string      `json:"ID"`
	Type                string      `json:"Type"`
	State               string      `json:"State,omitempty"`
	Flag                string      `json:"Flag,omitempty"`
	Enabled             *bool       `json:"Enabled,omitempty"`
	SwStatus            string      `json:"SoftwareStatus,omitempty"`
	Role                string      `json:"Role,omitempty"`
	SubRole             string      `json:"SubRole,omitempty"`
	NID                 json.Number `json:"NID,omitempty"`
	Subtype             string      `json:"Subtype,omitempty"`
	NetType             string      `json:"NetType,omitempty"`
	Arch                string      `json:"Arch,omitempty"`
	Class               string      `json:"Class,omitempty"`
	ReservationDisabled bool        `json:"ReservationDisabled,omitempty"`
	Locked              bool        `json:"Locked,omitempty"`
}

// ComponentStatus defines the observed state of Component
type ComponentStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
	Ready   bool   `json:"ready"`
	// Add your status fields here
}

// Validate implements custom validation logic for Component
func (r *Component) Validate(ctx context.Context) error {
	// Add custom validation logic here
	// Example:
	// if r.Spec.Name == "forbidden" {
	//     return errors.New("name 'forbidden' is not allowed")
	// }

	return nil
}

// GetKind returns the kind of the resource
func (r *Component) GetKind() string {
	return "Component"
}

// GetName returns the name of the resource
func (r *Component) GetName() string {
	return r.Metadata.Name
}

// GetUID returns the UID of the resource
func (r *Component) GetUID() string {
	return r.Metadata.UID
}

func init() {
	// Register resource type prefix for storage
	resource.RegisterResourcePrefix("Component", "com")
}
