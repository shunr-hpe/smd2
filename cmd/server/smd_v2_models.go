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
