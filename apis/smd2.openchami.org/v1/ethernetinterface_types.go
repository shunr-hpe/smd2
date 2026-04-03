package v1

import (
	"context"

	"github.com/openchami/fabrica/pkg/fabrica"
)

type EthernetInterface struct {
	APIVersion string                  `json:"apiVersion"`
	Kind       string                  `json:"kind"`
	Metadata   fabrica.Metadata        `json:"metadata"`
	ID         string                  `json:"id,omitempty"`
	Spec       EthernetInterfaceSpec   `json:"spec" validate:"required"`
	Status     EthernetInterfaceStatus `json:"status,omitempty"`
}

type EthernetInterfaceSpec struct {
	// todo resolve what to do about the case differences
	// SMD uses Description and fabrica uses description
	Description string      `json:"Description,omitempty" validate:"max=200"`
	ID          string      `json:"ID"`
	MACAddr     string      `json:"MACAddress"`
	LastUpdate  string      `json:"LastUpdate"`
	CompID      string      `json:"ComponentID"`
	Type        string      `json:"Type"`
	IPAddresses []IPAddress `json:"IPAddresses"`
}

type EthernetInterfaceStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
	Ready   bool   `json:"ready"`
}

func (r *EthernetInterface) Validate(ctx context.Context) error {

	return nil
}

func (r *EthernetInterface) GetKind() string {
	return "EthernetInterface"
}

func (r *EthernetInterface) GetName() string {
	return r.Metadata.Name
}

func (r *EthernetInterface) GetUID() string {
	return r.Metadata.UID
}

func (r *EthernetInterface) IsHub() {}

type IPAddress struct {
	IPAddress string `json:"IPAddress"`
	Network   string `json:"Network,omitempty"`
}
