// Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	v1 "github.com/OpenCHAMI/inventory-service/apis/inventory-service.openchami.org/v1"
	"github.com/OpenCHAMI/inventory-service/cmd/plugins"
	"github.com/go-chi/chi/v5"
	"github.com/openchami/fabrica/pkg/events"
	"github.com/openchami/fabrica/pkg/resource"
	"github.com/openchami/fabrica/pkg/validation"
	"github.com/openchami/fabrica/pkg/versioning"
)

// GetComponentEndpoint returns all Component resources
func GetComponentEndpointsCsm(w http.ResponseWriter, r *http.Request) {
	// Authorization: Add custom middleware in routes.go or implement checks here
	// Example: if !authorized(r) { respondError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized")); return }

	componentEndpoints, err := plugins.Store.LoadAllComponentEndpoints(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load componentendpoints: %w", err))
		return
	}
	componentEndpointArray := ComponentEndpointArray{
		ComponentEndpoints: make([]*v1.ComponentEndpointSpec, len(componentEndpoints)),
	}
	for i, c := range componentEndpoints {
		componentEndpointArray.ComponentEndpoints[i] = &c.Spec
	}
	respondJSON(w, http.StatusOK, componentEndpointArray)
}

// GetComponentEndpoint returns a specific Component resource by UID
func GetComponentEndpointCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("ComponentEndpoint ID is required"))
		return
	}

	// Version context available here for version-aware operations
	// versionCtx := versioning.GetVersionContext(r.Context())
	// Requested version: versionCtx.ServeVersion
	// To enable: replace plugins.Store.LoadComponentEndpoint() with version-aware function

	// Authorization: Add custom middleware in routes.go or implement checks here
	// Example: if !authorized(r) { respondError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized")); return }

	componentEndpoint, err := plugins.Store.LoadComponentEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("failed to find componentEndpoint %s: %w", id, err))
		return
	}

	if componentEndpoint == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("componentEndpoint not found: %s", id))
		return
	}
	respondJSON(w, http.StatusOK, &componentEndpoint.Spec)
}

// CreateComponentEndpoint creates a new ComponentEndpoint resource
func CreateComponentEndpointCsm(w http.ResponseWriter, r *http.Request) {
	var req ComponentEndpointArray
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	/* todo
	// Layer 1: Request validation (validates inline spec fields and metadata)
	if err := validation.ValidateResource(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
		return
	}
	*/

	// Get version context from request (set by version negotiation middleware)
	versionCtx := versioning.GetVersionContext(r.Context())
	for _, c := range req.ComponentEndpoints {
		uid, err := resource.GenerateUIDForResource("ComponentEndpoint")
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate UID: %w", err))
			return
		}

		// Versioned mode: flat fields with fabrica.Metadata
		componentEndpoint := &v1.ComponentEndpoint{
			// Use negotiated ServeVersion (from Accept header) for apiVersion
			APIVersion: versionCtx.ServeVersion,
			Kind:       "ComponentEndpoint",
			Spec:       *c,
		}
		// Initialize metadata from request
		componentEndpoint.Metadata.UID = uid
		componentEndpoint.Metadata.Name = c.ID
		now := time.Now()
		componentEndpoint.Metadata.CreatedAt = now
		componentEndpoint.Metadata.UpdatedAt = now

		// Set labels and annotations
		if componentEndpoint.Metadata.Labels == nil {
			componentEndpoint.Metadata.Labels = make(map[string]string)
		}
		if componentEndpoint.Metadata.Annotations == nil {
			componentEndpoint.Metadata.Annotations = make(map[string]string)
		}

		// Layer 2: Custom business logic validation
		if err := validation.ValidateWithContext(r.Context(), componentEndpoint); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
			return
		}
		// Set initial status
		// This assumes the generator passes an 'IsReconcilable' boolean
		// to this template, and that the resource has a .Status.Phase field.

		// Save (Layer 1: Ent validation happens automatically if using Ent storage)
		if err := plugins.Store.SaveComponentEndpoint(r.Context(), componentEndpoint); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save ComponentEndpoint: %w", err))
			return
		}

		// Publish resource created event

		if err := events.PublishResourceCreated(r.Context(), "ComponentEndpoint", componentEndpoint.Metadata.UID, componentEndpoint.Metadata.Name, componentEndpoint); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource created event for ComponentEndpoint %s: %v\n", componentEndpoint.Metadata.UID, err)
		}
	}

	respondJSON(w, http.StatusCreated, req)
}

// UpdateComponent updates the spec of an existing Component resource
// NOTE: This endpoint ONLY updates the spec. Use PUT //components/{uid}/status to update status.
func UpdateComponentEndpointCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("ComponentEndpoint ID is required"))
		return
	}
	componentEndpoint, err := plugins.Store.LoadComponentEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load componentEndpoint %s: %w", id, err))
		return
	}

	if componentEndpoint == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("component endpoint not found: %s", id))
		return
	}

	var req v1.ComponentEndpointSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Apply updates

	// Versioned mode: direct field access
	componentEndpoint.Metadata.Name = req.ID

	// Update spec fields ONLY - status should use /status subresource
	componentEndpoint.Spec = req

	// Update labels and annotations
	if componentEndpoint.Metadata.Labels == nil {
		componentEndpoint.Metadata.Labels = make(map[string]string)
	}
	if componentEndpoint.Metadata.Annotations == nil {
		componentEndpoint.Metadata.Annotations = make(map[string]string)
	}

	// Update timestamp
	componentEndpoint.Metadata.UpdatedAt = time.Now()

	if err := plugins.Store.SaveComponentEndpoint(r.Context(), componentEndpoint); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save ComponentEndpoint: %w", err))
		return
	}

	// Publish resource updated event
	updateMetadata := map[string]interface{}{
		"updatedAt": componentEndpoint.Metadata.UpdatedAt,
	}

	if err := events.PublishResourceUpdated(r.Context(), "ComponentEndpoint", componentEndpoint.Metadata.UID, componentEndpoint.Metadata.Name, componentEndpoint, updateMetadata); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource updated event for ComponentEndpoint %s: %v\n", componentEndpoint.Metadata.UID, err)

	}

	respondJSON(w, http.StatusOK, componentEndpoint.Spec)
}

// DeleteComponent deletes a Component resource
func DeleteComponentEndpointCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Component Endpoint ID is required"))
		return
	}

	componentEndpoint, err := plugins.Store.LoadComponentEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("failed to find componentEndpoint %s: %w", id, err))
		return
	}

	if componentEndpoint != nil {
		uid := componentEndpoint.GetUID()
		if err := plugins.Store.DeleteComponentEndpoint(r.Context(), uid); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete ComponentEndpoint: %w", err))
			return
		}
		// Publish resource deleted event
		deleteMetadata := map[string]interface{}{
			"deletedAt": time.Now(),
		}

		if err := events.PublishResourceDeleted(r.Context(), "ComponentEndpoint", componentEndpoint.Metadata.UID, componentEndpoint.Metadata.Name, deleteMetadata); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource deleted event for ComponentEndpoint %s: %v\n", componentEndpoint.Metadata.UID, err)

		}
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "ComponentEndpoint deleted successfully",
			UID:     uid,
		})
	} else {
		// todo maybe do something different
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "ComponentEndpoint not present",
			UID:     "",
		})
	}
}
