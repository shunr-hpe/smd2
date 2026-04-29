// Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package plugins

import (
	"context"

	v1 "github.com/OpenCHAMI/inventory-service/apis/inventory-service.openchami.org/v1"
)

var Store Storage

// Storage defines the interface for all storage operations.
type Storage interface {
	LoadAllComponents(ctx context.Context) ([]*v1.Component, error)
	LoadComponent(ctx context.Context, uid string) (*v1.Component, error)
	SaveComponent(ctx context.Context, resource *v1.Component) error
	DeleteComponent(ctx context.Context, uid string) error

	LoadAllComponentEndpoints(ctx context.Context) ([]*v1.ComponentEndpoint, error)
	LoadComponentEndpoint(ctx context.Context, uid string) (*v1.ComponentEndpoint, error)
	SaveComponentEndpoint(ctx context.Context, resource *v1.ComponentEndpoint) error
	DeleteComponentEndpoint(ctx context.Context, uid string) error

	LoadAllEthernetInterfaces(ctx context.Context) ([]*v1.EthernetInterface, error)
	LoadEthernetInterface(ctx context.Context, uid string) (*v1.EthernetInterface, error)
	SaveEthernetInterface(ctx context.Context, resource *v1.EthernetInterface) error
	DeleteEthernetInterface(ctx context.Context, uid string) error

	LoadAllGroups(ctx context.Context) ([]*v1.Group, error)
	LoadGroup(ctx context.Context, uid string) (*v1.Group, error)
	SaveGroup(ctx context.Context, resource *v1.Group) error
	DeleteGroup(ctx context.Context, uid string) error

	LoadAllRedfishEndpoints(ctx context.Context) ([]*v1.RedfishEndpoint, error)
	LoadRedfishEndpoint(ctx context.Context, uid string) (*v1.RedfishEndpoint, error)
	SaveRedfishEndpoint(ctx context.Context, resource *v1.RedfishEndpoint) error
	DeleteRedfishEndpoint(ctx context.Context, uid string) error

	LoadAllHardwares(ctx context.Context) ([]*v1.Hardware, error)
	LoadHardware(ctx context.Context, uid string) (*v1.Hardware, error)
	SaveHardware(ctx context.Context, resource *v1.Hardware) error
	DeleteHardware(ctx context.Context, uid string) error

	LoadAllServiceEndpoints(ctx context.Context) ([]*v1.ServiceEndpoint, error)
	LoadServiceEndpoint(ctx context.Context, uid string) (*v1.ServiceEndpoint, error)
	SaveServiceEndpoint(ctx context.Context, resource *v1.ServiceEndpoint) error
	DeleteServiceEndpoint(ctx context.Context, uid string) error

	LoadComponentByID(ctx context.Context, id string) (*v1.Component, error)
	LoadComponentEndpointByID(ctx context.Context, id string) (*v1.ComponentEndpoint, error)
	LoadRedfishEndpointByID(ctx context.Context, id string) (*v1.RedfishEndpoint, error)
	LoadEthernetInterfaceByID(ctx context.Context, id string) (*v1.EthernetInterface, error)
	LoadServiceEndpointByID(ctx context.Context, id string) (*v1.ServiceEndpoint, error)
	LoadGroupByLabel(ctx context.Context, label string) (*v1.Group, error)

	StorageExtras
}
