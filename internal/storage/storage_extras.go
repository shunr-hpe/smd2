// Copyright © 2025-2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package storage

import (
	"context"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	v1 "github.com/OpenCHAMI/smd2/apis/smd2.openchami.org/v1"
	"github.com/OpenCHAMI/smd2/internal/storage/ent"
	entpredicate "github.com/OpenCHAMI/smd2/internal/storage/ent/predicate"
	entresource "github.com/OpenCHAMI/smd2/internal/storage/ent/resource"
)

// LoadComponentByID loads a single Component resource by its Spec.ID from Ent storage
func LoadComponentByID(ctx context.Context, id string) (*v1.Component, error) {
	if entClient == nil {
		return nil, fmt.Errorf("ent client not initialized")
	}

	entResource, err := entClient.Resource.Query().
		Where(
			entresource.KindEQ("Component"),
			entresource.ResourceIDEQ(id),
		).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load Component with ID %s: %w", id, err)
	}

	// Convert to Fabrica resource
	fabricaResource, err := FromEntResource(ctx, entResource)
	if err != nil {
		return nil, err
	}

	return fabricaResource.(*v1.Component), nil
}

func LoadComponentEndpointByID(ctx context.Context, id string) (*v1.ComponentEndpoint, error) {
	if entClient == nil {
		return nil, fmt.Errorf("ent client not initialized")
	}

	entResource, err := entClient.Resource.Query().
		Where(
			entresource.KindEQ("ComponentEndpoint"),
			entresource.ResourceIDEQ(id),
		).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load ComponentEndpoint with ID %s: %w", id, err)
	}

	fabricaResource, err := FromEntResource(ctx, entResource)
	if err != nil {
		return nil, err
	}

	return fabricaResource.(*v1.ComponentEndpoint), nil
}

func LoadRedfishEndpointByID(ctx context.Context, id string) (*v1.RedfishEndpoint, error) {
	if entClient == nil {
		return nil, fmt.Errorf("ent client not initialized")
	}

	entResource, err := entClient.Resource.Query().
		Where(
			entresource.KindEQ("RedfishEndpoint"),
			entresource.ResourceIDEQ(id),
		).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load RedfishEndpoint with ID %s: %w", id, err)
	}

	fabricaResource, err := FromEntResource(ctx, entResource)
	if err != nil {
		return nil, err
	}

	return fabricaResource.(*v1.RedfishEndpoint), nil
}

func LoadEthernetInterfaceByID(ctx context.Context, id string) (*v1.EthernetInterface, error) {
	if entClient == nil {
		return nil, fmt.Errorf("ent client not initialized")
	}

	entResource, err := entClient.Resource.Query().
		Where(
			entresource.KindEQ("EthernetInterface"),
			entresource.ResourceIDEQ(id),
		).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load EthernetInterface with ID %s: %w", id, err)
	}

	fabricaResource, err := FromEntResource(ctx, entResource)
	if err != nil {
		return nil, err
	}

	return fabricaResource.(*v1.EthernetInterface), nil
}

func LoadServiceEndpointByID(ctx context.Context, id string) (*v1.ServiceEndpoint, error) {
	if entClient == nil {
		return nil, fmt.Errorf("ent client not initialized")
	}

	entResource, err := entClient.Resource.Query().
		Where(
			entresource.KindEQ("ServiceEndpoint"),
			entresource.ResourceIDEQ(id),
		).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load ServiceEndpoint with ID %s: %w", id, err)
	}

	fabricaResource, err := FromEntResource(ctx, entResource)
	if err != nil {
		return nil, err
	}

	return fabricaResource.(*v1.ServiceEndpoint), nil
}

// LoadServiceEndpointsByServiceID loads all ServiceEndpoint resources whose
// spec.RedfishType matches the given serviceID.
func LoadServiceEndpointsByRedfishType(ctx context.Context, serviceID string) ([]*v1.ServiceEndpoint, error) {
	if entClient == nil {
		return nil, fmt.Errorf("ent client not initialized")
	}

	jsonPredicate := entpredicate.Resource(func(s *entsql.Selector) {
		s.Where(sqljson.ValueEQ(s.C(entresource.FieldSpec), serviceID,
			sqljson.Path("RedfishType")))
	})

	entResources, err := entClient.Resource.Query().
		Where(
			entresource.KindEQ("ServiceEndpoint"),
			jsonPredicate,
		).
		WithLabels().
		WithAnnotations().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load ServiceEndpoint resources: %w", err)
	}

	var results []*v1.ServiceEndpoint
	for _, entResource := range entResources {
		fabricaResource, err := FromEntResource(ctx, entResource)
		if err != nil {
			continue
		}
		results = append(results, fabricaResource.(*v1.ServiceEndpoint))
	}

	return results, nil
}

func LoadGroupByLabel(ctx context.Context, label string) (*v1.Group, error) {
	if entClient == nil {
		return nil, fmt.Errorf("ent client not initialized")
	}

	entResource, err := entClient.Resource.Query().
		Where(
			entresource.KindEQ("Group"),
			entresource.ResourceIDEQ(label),
		).
		WithLabels().
		WithAnnotations().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to load Group with label %s: %w", label, err)
	}

	fabricaResource, err := FromEntResource(ctx, entResource)
	if err != nil {
		return nil, err
	}

	return fabricaResource.(*v1.Group), nil
}
