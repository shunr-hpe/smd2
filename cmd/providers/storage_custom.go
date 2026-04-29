// Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package providers

import (
	"context"

	v1 "github.com/OpenCHAMI/inventory-service/apis/inventory-service.openchami.org/v1"
	"github.com/OpenCHAMI/inventory-service/internal/storage"
)

// EntStorage implements the plugins.Storage interface by delegating to the
// package-level functions in internal/storage.
type EntStorage struct{}

func (s *EntStorage) LoadAllComponents(ctx context.Context) ([]*v1.Component, error) {
	return storage.LoadAllComponents(ctx)
}

func (s *EntStorage) LoadComponent(ctx context.Context, uid string) (*v1.Component, error) {
	return storage.LoadComponent(ctx, uid)
}

func (s *EntStorage) SaveComponent(ctx context.Context, resource *v1.Component) error {
	resource.ID = resource.Spec.ID
	return storage.SaveComponent(ctx, resource)
}

func (s *EntStorage) DeleteComponent(ctx context.Context, uid string) error {
	return storage.DeleteComponent(ctx, uid)
}

func (s *EntStorage) LoadAllComponentEndpoints(ctx context.Context) ([]*v1.ComponentEndpoint, error) {
	return storage.LoadAllComponentEndpoints(ctx)
}

func (s *EntStorage) LoadComponentEndpoint(ctx context.Context, uid string) (*v1.ComponentEndpoint, error) {
	return storage.LoadComponentEndpoint(ctx, uid)
}

func (s *EntStorage) SaveComponentEndpoint(ctx context.Context, resource *v1.ComponentEndpoint) error {
	resource.ID = resource.Spec.ID
	return storage.SaveComponentEndpoint(ctx, resource)
}

func (s *EntStorage) DeleteComponentEndpoint(ctx context.Context, uid string) error {
	return storage.DeleteComponentEndpoint(ctx, uid)
}

func (s *EntStorage) LoadAllEthernetInterfaces(ctx context.Context) ([]*v1.EthernetInterface, error) {
	return storage.LoadAllEthernetInterfaces(ctx)
}

func (s *EntStorage) LoadEthernetInterface(ctx context.Context, uid string) (*v1.EthernetInterface, error) {
	return storage.LoadEthernetInterface(ctx, uid)
}

func (s *EntStorage) SaveEthernetInterface(ctx context.Context, resource *v1.EthernetInterface) error {
	resource.ID = resource.Spec.ID
	return storage.SaveEthernetInterface(ctx, resource)
}

func (s *EntStorage) DeleteEthernetInterface(ctx context.Context, uid string) error {
	return storage.DeleteEthernetInterface(ctx, uid)
}

func (s *EntStorage) LoadAllGroups(ctx context.Context) ([]*v1.Group, error) {
	return storage.LoadAllGroups(ctx)
}

func (s *EntStorage) LoadGroup(ctx context.Context, uid string) (*v1.Group, error) {
	return storage.LoadGroup(ctx, uid)
}

func (s *EntStorage) SaveGroup(ctx context.Context, resource *v1.Group) error {
	resource.ID = resource.Spec.Label
	return storage.SaveGroup(ctx, resource)
}

func (s *EntStorage) DeleteGroup(ctx context.Context, uid string) error {
	return storage.DeleteGroup(ctx, uid)
}

func (s *EntStorage) LoadAllHardwares(ctx context.Context) ([]*v1.Hardware, error) {
	return storage.LoadAllHardwares(ctx)
}

func (s *EntStorage) LoadHardware(ctx context.Context, uid string) (*v1.Hardware, error) {
	return storage.LoadHardware(ctx, uid)
}

func (s *EntStorage) SaveHardware(ctx context.Context, resource *v1.Hardware) error {
	resource.ID = resource.Spec.ID
	return storage.SaveHardware(ctx, resource)
}

func (s *EntStorage) DeleteHardware(ctx context.Context, uid string) error {
	return storage.DeleteHardware(ctx, uid)
}

func (s *EntStorage) LoadAllRedfishEndpoints(ctx context.Context) ([]*v1.RedfishEndpoint, error) {
	return storage.LoadAllRedfishEndpoints(ctx)
}

func (s *EntStorage) LoadRedfishEndpoint(ctx context.Context, uid string) (*v1.RedfishEndpoint, error) {
	return storage.LoadRedfishEndpoint(ctx, uid)
}

func (s *EntStorage) SaveRedfishEndpoint(ctx context.Context, resource *v1.RedfishEndpoint) error {
	resource.ID = resource.Spec.ID
	return storage.SaveRedfishEndpoint(ctx, resource)
}

func (s *EntStorage) DeleteRedfishEndpoint(ctx context.Context, uid string) error {
	return storage.DeleteRedfishEndpoint(ctx, uid)
}

func (s *EntStorage) LoadAllServiceEndpoints(ctx context.Context) ([]*v1.ServiceEndpoint, error) {
	return storage.LoadAllServiceEndpoints(ctx)
}

func (s *EntStorage) LoadServiceEndpoint(ctx context.Context, uid string) (*v1.ServiceEndpoint, error) {
	return storage.LoadServiceEndpoint(ctx, uid)
}

func (s *EntStorage) SaveServiceEndpoint(ctx context.Context, resource *v1.ServiceEndpoint) error {
	resource.ID = resource.Spec.RedfishType + "-" + resource.Spec.RfEndpointID
	return storage.SaveServiceEndpoint(ctx, resource)
}

func (s *EntStorage) DeleteServiceEndpoint(ctx context.Context, uid string) error {
	return storage.DeleteServiceEndpoint(ctx, uid)
}

func (s *EntStorage) LoadComponentByID(ctx context.Context, id string) (*v1.Component, error) {
	return storage.LoadComponentByID(ctx, id)
}

func (s *EntStorage) LoadComponentEndpointByID(ctx context.Context, id string) (*v1.ComponentEndpoint, error) {
	return storage.LoadComponentEndpointByID(ctx, id)
}

func (s *EntStorage) LoadRedfishEndpointByID(ctx context.Context, id string) (*v1.RedfishEndpoint, error) {
	return storage.LoadRedfishEndpointByID(ctx, id)
}

func (s *EntStorage) LoadEthernetInterfaceByID(ctx context.Context, id string) (*v1.EthernetInterface, error) {
	return storage.LoadEthernetInterfaceByID(ctx, id)
}

func (s *EntStorage) LoadServiceEndpointByID(ctx context.Context, id string) (*v1.ServiceEndpoint, error) {
	return storage.LoadServiceEndpointByID(ctx, id)
}

func (s *EntStorage) LoadGroupByLabel(ctx context.Context, label string) (*v1.Group, error) {
	return storage.LoadGroupByLabel(ctx, label)
}

func (s *EntStorage) LoadHardwareByID(ctx context.Context, id string) (*v1.Hardware, error) {
	return storage.LoadHardwareByID(ctx, id)
}

func (s *EntStorage) LoadServiceEndpointsByRedfishType(ctx context.Context, redfishType string) ([]*v1.ServiceEndpoint, error) {
	return storage.LoadServiceEndpointsByRedfishType(ctx, redfishType)
}

func (s *EntStorage) LoadServiceEndpointsByRedfishTypeAndID(ctx context.Context, redfishType string, redfishID string) ([]*v1.ServiceEndpoint, error) {
	return storage.LoadServiceEndpointsByRedfishTypeAndID(ctx, redfishType, redfishID)
}
