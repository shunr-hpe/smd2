// Copyright © 2025-2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package storage

import (
	"context"
	"fmt"

	v1 "github.com/OpenCHAMI/smd2/apis/smd2.openchami.org/v1"
)

func LoadComponentByID(ctx context.Context, id string) (*v1.Component, error) {
	// todo Change to not have to read all components
	components, err := LoadAllComponents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load components: %w", err)
	}
	var component *v1.Component
	for _, c := range components {
		if c.Spec.ID == id {
			component = c
			break
		}
	}
	return component, nil
}

func LoadComponentEndpointByID(ctx context.Context, id string) (*v1.ComponentEndpoint, error) {
	// todo Change to not have to read all component endpoints
	componentEndpoints, err := LoadAllComponentEndpoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load component endpoints: %w", err)
	}
	var componentEndpoint *v1.ComponentEndpoint
	for _, c := range componentEndpoints {
		if c.Spec.ID == id {
			componentEndpoint = c
			break
		}
	}
	return componentEndpoint, nil
}

func LoadRedfishEndpointByID(ctx context.Context, id string) (*v1.RedfishEndpoint, error) {
	// todo Change to not have to read all redfish endpoints
	redfishEndpoints, err := LoadAllRedfishEndpoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load redfish endpoints: %w", err)
	}
	var redfishEndpoint *v1.RedfishEndpoint
	for _, re := range redfishEndpoints {
		if re.Spec.ID == id {
			redfishEndpoint = re
			break
		}
	}
	return redfishEndpoint, nil
}

func LoadEthernetInterfaceByID(ctx context.Context, id string) (*v1.EthernetInterface, error) {
	// todo Change to not have to read all ethernet interfaces
	ethernetInterfaces, err := LoadAllEthernetInterfaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load ethernet interfaces: %w", err)
	}
	var ethernetInterface *v1.EthernetInterface
	for _, e := range ethernetInterfaces {
		if e.Spec.ID == id {
			ethernetInterface = e
			break
		}
	}
	return ethernetInterface, nil
}

func LoadServiceEndpointByID(ctx context.Context, id string) (*v1.ServiceEndpoint, error) {
	// todo Change to not have to read all service endpoints
	serviceEndpoints, err := LoadAllServiceEndpoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load service endpoints: %w", err)
	}
	var serviceEndpoint *v1.ServiceEndpoint
	for _, s := range serviceEndpoints {
		if s.Spec.RfEndpointID == id {
			serviceEndpoint = s
			break
		}
	}
	return serviceEndpoint, nil
}

func LoadGroupByLabel(ctx context.Context, label string) (*v1.Group, error) {
	// todo Change to not have to read all groups
	groups, err := LoadAllGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load groups: %w", err)
	}
	var group *v1.Group
	for _, g := range groups {
		if g.Spec.Label == label {
			group = g
			break
		}
	}
	return group, nil
}
