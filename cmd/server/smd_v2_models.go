// Copyright © 2025-2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
package main

import (
	v1 "github.com/OpenCHAMI/smd2/apis/smd2.openchami.org/v1"
)

type ComponentArray struct {
	Components []*v1.ComponentSpec `json:"Components"`
}

type ComponentEndpointArray struct {
	ComponentEndpoints []*v1.ComponentEndpointSpec `json:"ComponentEndpoints"`
}

type EthernetInterfaceArray struct {
	EthernetInterfaces []*v1.EthernetInterfaceSpec `json:"EthernetInterfaces"`
}

type ServiceEndpointArray struct {
	ServiceEndpoints []*v1.ServiceEndpointSpec `json:"ServiceEndpoints"`
}

type RedfishEndpointArray struct {
	RedfishEndpoints []*v1.RedfishEndpointSpec `json:"RedfishEndpoints"`
}

// RedfishEndpointV2EthernetInterface is a network interface entry in the V2 inventory format
// (mirrors schemas.EthernetInterface from OpenCHAMI/smd).
type RedfishEndpointV2EthernetInterface struct {
	URI         string `json:"uri,omitempty"`
	MAC         string `json:"mac,omitempty"`
	IP          string `json:"ip,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
}

// RedfishEndpointV2Manager is a manager (BMC) entry in the V2 inventory format.
type RedfishEndpointV2Manager struct {
	URI                string                               `json:"uri,omitempty"`
	UUID               string                               `json:"uuid,omitempty"`
	Name               string                               `json:"name,omitempty"`
	Description        string                               `json:"description,omitempty"`
	Model              string                               `json:"model,omitempty"`
	Type               string                               `json:"type,omitempty"`
	FirmwareVersion    string                               `json:"firmware_version,omitempty"`
	EthernetInterfaces []RedfishEndpointV2EthernetInterface `json:"ethernet_interfaces,omitempty"`
}

// RedfishEndpointV2System is a compute system entry in the V2 inventory format
// (mirrors schemas.InventoryDetail from OpenCHAMI/smd).
type RedfishEndpointV2System struct {
	URI                string                               `json:"uri,omitempty"`
	UUID               string                               `json:"uuid,omitempty"`
	Manufacturer       string                               `json:"manufacturer,omitempty"`
	SystemType         string                               `json:"system_type,omitempty"`
	Name               string                               `json:"name,omitempty"`
	Model              string                               `json:"model,omitempty"`
	Serial             string                               `json:"serial,omitempty"`
	BiosVersion        string                               `json:"bios_version,omitempty"`
	EthernetInterfaces []RedfishEndpointV2EthernetInterface `json:"ethernet_interfaces,omitempty"`
}

// RedfishEndpointV2Request is the V2 POST body format used by parseRedfishEndpointDataV2
// in OpenCHAMI/smd. It embeds the standard RedfishEndpointSpec fields (ID, FQDN, etc.)
// alongside Systems and Managers inventory arrays.
// Presence of non-empty Systems or Managers indicates a V2 request.
type RedfishEndpointV2Request struct {
	v1.RedfishEndpointSpec
	Systems  []RedfishEndpointV2System   `json:"Systems,omitempty"`
	Managers []RedfishEndpointV2Manager  `json:"Managers,omitempty"`
}
