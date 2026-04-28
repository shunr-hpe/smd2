package storage

import (
	"context"
	"fmt"

	"github.com/OpenCHAMI/inventory-service/internal/storage/ent"
	"github.com/OpenCHAMI/inventory-service/internal/storage/ent/label"
	entresource "github.com/OpenCHAMI/inventory-service/internal/storage/ent/resource"

	v1 "github.com/OpenCHAMI/inventory-service/apis/inventory-service.openchami.org/v1"
)

// ensureEntClient verifies the ent client has been initialized
func ensureEntClient() {
	if entClient == nil {
		panic("ent client not initialized: call SetEntClient in main.go before using storage")
	}
}

// QueryResources returns a generic query builder for a given kind
func QueryResources(ctx context.Context, kind string) *ent.ResourceQuery {
	ensureEntClient()
	return entClient.Resource.Query().
		Where(entresource.KindEQ(kind))
}

// QueryResourcesByLabels queries resources by kind and exact-match labels
func QueryResourcesByLabels(ctx context.Context, kind string, labels map[string]string) (*ent.ResourceQuery, error) {
	ensureEntClient()
	q := entClient.Resource.Query().Where(entresource.KindEQ(kind))
	for k, v := range labels {
		q = q.Where(entresource.HasLabelsWith(
			label.KeyEQ(k),
			label.ValueEQ(v),
		))
	}
	return q, nil
}

// Querycomponents returns a query builder for components
func Querycomponents(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "Component")
}

// GetComponentByUID loads a single Component by UID
func GetComponentByUID(ctx context.Context, uid string) (*v1.Component, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("Component")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load Component %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.Component), nil
}

// ListcomponentsByLabels returns components matching all provided labels
func ListcomponentsByLabels(ctx context.Context, labels map[string]string) ([]*v1.Component, error) {
	q, err := QueryResourcesByLabels(ctx, "Component", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.Component, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.Component))
	}
	return out, nil
}

// Querycomponentendpoints returns a query builder for componentendpoints
func Querycomponentendpoints(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "ComponentEndpoint")
}

// GetComponentEndpointByUID loads a single ComponentEndpoint by UID
func GetComponentEndpointByUID(ctx context.Context, uid string) (*v1.ComponentEndpoint, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("ComponentEndpoint")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load ComponentEndpoint %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.ComponentEndpoint), nil
}

// ListcomponentendpointsByLabels returns componentendpoints matching all provided labels
func ListcomponentendpointsByLabels(ctx context.Context, labels map[string]string) ([]*v1.ComponentEndpoint, error) {
	q, err := QueryResourcesByLabels(ctx, "ComponentEndpoint", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ComponentEndpoint, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.ComponentEndpoint))
	}
	return out, nil
}

// Queryethernetinterfaces returns a query builder for ethernetinterfaces
func Queryethernetinterfaces(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "EthernetInterface")
}

// GetEthernetInterfaceByUID loads a single EthernetInterface by UID
func GetEthernetInterfaceByUID(ctx context.Context, uid string) (*v1.EthernetInterface, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("EthernetInterface")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load EthernetInterface %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.EthernetInterface), nil
}

// ListethernetinterfacesByLabels returns ethernetinterfaces matching all provided labels
func ListethernetinterfacesByLabels(ctx context.Context, labels map[string]string) ([]*v1.EthernetInterface, error) {
	q, err := QueryResourcesByLabels(ctx, "EthernetInterface", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.EthernetInterface, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.EthernetInterface))
	}
	return out, nil
}

// Querygroups returns a query builder for groups
func Querygroups(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "Group")
}

// GetGroupByUID loads a single Group by UID
func GetGroupByUID(ctx context.Context, uid string) (*v1.Group, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("Group")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load Group %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.Group), nil
}

// ListgroupsByLabels returns groups matching all provided labels
func ListgroupsByLabels(ctx context.Context, labels map[string]string) ([]*v1.Group, error) {
	q, err := QueryResourcesByLabels(ctx, "Group", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.Group, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.Group))
	}
	return out, nil
}

// Queryhardwares returns a query builder for hardwares
func Queryhardwares(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "Hardware")
}

// GetHardwareByUID loads a single Hardware by UID
func GetHardwareByUID(ctx context.Context, uid string) (*v1.Hardware, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("Hardware")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load Hardware %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.Hardware), nil
}

// ListhardwaresByLabels returns hardwares matching all provided labels
func ListhardwaresByLabels(ctx context.Context, labels map[string]string) ([]*v1.Hardware, error) {
	q, err := QueryResourcesByLabels(ctx, "Hardware", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.Hardware, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.Hardware))
	}
	return out, nil
}

// Queryredfishendpoints returns a query builder for redfishendpoints
func Queryredfishendpoints(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "RedfishEndpoint")
}

// GetRedfishEndpointByUID loads a single RedfishEndpoint by UID
func GetRedfishEndpointByUID(ctx context.Context, uid string) (*v1.RedfishEndpoint, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("RedfishEndpoint")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load RedfishEndpoint %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.RedfishEndpoint), nil
}

// ListredfishendpointsByLabels returns redfishendpoints matching all provided labels
func ListredfishendpointsByLabels(ctx context.Context, labels map[string]string) ([]*v1.RedfishEndpoint, error) {
	q, err := QueryResourcesByLabels(ctx, "RedfishEndpoint", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.RedfishEndpoint, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.RedfishEndpoint))
	}
	return out, nil
}

// Queryserviceendpoints returns a query builder for serviceendpoints
func Queryserviceendpoints(ctx context.Context) *ent.ResourceQuery {
	return QueryResources(ctx, "ServiceEndpoint")
}

// GetServiceEndpointByUID loads a single ServiceEndpoint by UID
func GetServiceEndpointByUID(ctx context.Context, uid string) (*v1.ServiceEndpoint, error) {
	ensureEntClient()
	r, err := entClient.Resource.Query().
		Where(entresource.UIDEQ(uid), entresource.KindEQ("ServiceEndpoint")).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load ServiceEndpoint %s: %w", uid, err)
	}
	v, err := FromEntResource(ctx, r)
	if err != nil {
		return nil, err
	}
	return v.(*v1.ServiceEndpoint), nil
}

// ListserviceendpointsByLabels returns serviceendpoints matching all provided labels
func ListserviceendpointsByLabels(ctx context.Context, labels map[string]string) ([]*v1.ServiceEndpoint, error) {
	q, err := QueryResourcesByLabels(ctx, "ServiceEndpoint", labels)
	if err != nil {
		return nil, err
	}
	rs, err := q.WithLabels().WithAnnotations().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ServiceEndpoint, 0, len(rs))
	for _, r := range rs {
		v, err := FromEntResource(ctx, r)
		if err != nil {
			continue
		}
		out = append(out, v.(*v1.ServiceEndpoint))
	}
	return out, nil
}
