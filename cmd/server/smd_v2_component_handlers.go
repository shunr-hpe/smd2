// Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/openchami/fabrica/pkg/events"
	"github.com/openchami/fabrica/pkg/resource"
	"github.com/openchami/fabrica/pkg/validation"
	"github.com/openchami/fabrica/pkg/versioning"
	"github.com/user/smd2/internal/storage"
	"github.com/user/smd2/pkg/resources/component"
)

// GetComponents returns all Component resources
func GetComponentsSmdV2(w http.ResponseWriter, r *http.Request) {
	// Authorization: Add custom middleware in routes.go or implement checks here
	// Example: if !authorized(r) { respondError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized")); return }

	components, err := storage.LoadAllComponents(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load components: %w", err))
		return
	}
	componentsSmdV2 := ComponentArray{
		Components: make([]*component.ComponentSpec, len(components)),
	}
	for i, c := range components {
		componentsSmdV2.Components[i] = &c.Spec
	}
	respondJSON(w, http.StatusOK, componentsSmdV2)
}

// GetComponent returns a specific Component resource by UID
func GetComponentSmdV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("component id is required"))
		return
	}

	// Version context available here for version-aware operations
	// versionCtx := versioning.GetVersionContext(r.Context())
	// Requested version: versionCtx.ServeVersion
	// To enable: replace storage.LoadComponent() with version-aware function

	// Authorization: Add custom middleware in routes.go or implement checks here
	// Example: if !authorized(r) { respondError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized")); return }

	components, err := storage.LoadAllComponents(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load components: %w", err))
		return
	}
	var component *component.ComponentSpec
	for _, c := range components {
		if c.Spec.ID == id {
			component = &c.Spec
			break
		}
	}

	if component == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("component not found: %s", id))
		return
	}
	respondJSON(w, http.StatusOK, component)
}

// CreateComponent creates a new Component resource
func CreateComponentSmdV2(w http.ResponseWriter, r *http.Request) {
	var req ComponentArray
	// var req CreateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	for _, c := range req.Components {
		// Get version context from request
		versionCtx := versioning.GetVersionContext(r.Context())

		uid, err := resource.GenerateUIDForResource("Component")
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate UID: %w", err))
			return
		}

		component := &component.Component{
			Resource: resource.Resource{
				APIVersion:    versionCtx.GroupVersion,
				Kind:          "Component",
				SchemaVersion: versionCtx.ServeVersion,
			},
			Spec: *c,
		}

		component.Metadata.Initialize(c.ID, uid)

		// Layer 2: Fabrica struct tag validation
		if err := validation.ValidateResource(component); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
			return
		}

		// Layer 3: Custom business logic validation
		if err := validation.ValidateWithContext(r.Context(), component); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
			return
		}

		// Set initial status

		// Save (Layer 1: Ent validation happens automatically if using Ent storage)
		if err := storage.SaveComponent(r.Context(), component); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save Component: %w", err))
			return
		}

		// Publish resource created event
		if err := events.PublishResourceCreated(r.Context(), "Component", component.GetUID(), component.GetName(), component); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource created event for Component %s: %v\n", component.GetUID(), err)
		}
	}

	// respondJSON(w, http.StatusCreated, component)
	respondJSON(w, http.StatusCreated, nil)
}

// UpdateComponent updates the spec of an existing Component resource
// NOTE: This endpoint ONLY updates the spec. Use PUT //components/{uid}/status to update status.
func UpdateComponentSmdV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Component ID is required"))
		return
	}

	components, err := storage.LoadAllComponents(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load components: %w", err))
		return
	}
	var component *component.Component
	for _, c := range components {
		if c.Spec.ID == id {
			component = c
			break
		}
	}

	if component == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("component not found: %s", id))
		return
	}

	var req UpdateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Apply updates
	if req.Name != "" {
		component.SetName(req.Name)
	}

	// Update spec fields ONLY - status should use /status subresource
	component.Spec = req.ComponentSpec

	// Update labels and annotations
	for k, v := range req.Labels {
		component.SetLabel(k, v)
	}
	for k, v := range req.Annotations {
		component.SetAnnotation(k, v)
	}

	component.Touch()

	if err := storage.SaveComponent(r.Context(), component); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save Component: %w", err))
		return
	}

	// Publish resource updated event
	updateMetadata := map[string]interface{}{
		"updatedAt": component.Metadata.UpdatedAt,
	}
	if err := events.PublishResourceUpdated(r.Context(), "Component", component.GetUID(), component.GetName(), component, updateMetadata); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource updated event for Component %s: %v\n", component.GetUID(), err)
	}

	respondJSON(w, http.StatusOK, component.Spec)
}

// DeleteComponent deletes a Component resource
func DeleteComponentSmdV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Component ID is required"))
		return
	}

	components, err := storage.LoadAllComponents(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load components: %w", err))
		return
	}
	var component *component.Component
	for _, c := range components {
		if c.Spec.ID == id {
			component = c
			break
		}
	}

	if component != nil {
		uid := component.GetUID()

		if err := storage.DeleteComponent(r.Context(), uid); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete Component: %w", err))
			return
		}

		// Publish resource deleted event
		deleteMetadata := map[string]interface{}{
			"deletedAt": time.Now(),
		}
		if err := events.PublishResourceDeleted(r.Context(), "Component", component.GetUID(), component.GetName(), deleteMetadata); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource deleted event for Component %s: %v\n", component.GetUID(), err)
		}
	}

	respondJSON(w, http.StatusOK, &struct {
		Message string
		ID      string
	}{
		Message: "Component deleted successfully",
		ID:      id,
	})
}
