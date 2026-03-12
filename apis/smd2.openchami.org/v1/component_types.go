package v1

import (
	"context"
	"encoding/json"

	"github.com/openchami/fabrica/pkg/fabrica"
)

type Component struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   fabrica.Metadata `json:"metadata"`
	ID         string           `json:"id,omitempty"`
	Spec       ComponentSpec    `json:"spec" validate:"required"`
	Status     ComponentStatus  `json:"status,omitempty"`
}

type ComponentSpec struct {
	Description string `json:"description,omitempty" validate:"max=200"`
	ID          string `json:"ID"`

	Type     string `json:"Type"`
	State    string `json:"State,omitempty"`
	Flag     string `json:"Flag,omitempty"`
	Enabled  *bool  `json:"Enabled,omitempty"`
	SwStatus string `json:"SoftwareStatus,omitempty"`
	Role     string `json:"Role,omitempty"`
	SubRole  string `json:"SubRole,omitempty"`

	NID     json.Number `json:"NID,omitempty"`
	Subtype string      `json:"Subtype,omitempty"`
	NetType string      `json:"NetType,omitempty"`

	Arch                string `json:"Arch,omitempty"`
	Class               string `json:"Class,omitempty"`
	ReservationDisabled bool   `json:"ReservationDisabled,omitempty"`
	Locked              bool   `json:"Locked,omitempty"`
}

type ComponentStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
	Ready   bool   `json:"ready"`
}

func (r *Component) Validate(ctx context.Context) error {

	return nil
}

func (r *Component) GetKind() string {
	return "Component"
}

func (r *Component) GetName() string {
	return r.Metadata.Name
}

func (r *Component) GetUID() string {
	return r.Metadata.UID
}

func (r *Component) IsHub() {}
