// Copyright © 2025-2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	v1 "github.com/OpenCHAMI/smd2/apis/smd2.openchami.org/v1"
	"github.com/OpenCHAMI/smd2/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/openchami/fabrica/pkg/events"
	"github.com/openchami/fabrica/pkg/resource"
	"github.com/openchami/fabrica/pkg/validation"
	"github.com/openchami/fabrica/pkg/versioning"
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
		Components: make([]*v1.ComponentSpec, len(components)),
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

	component, err := storage.LoadComponentByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to component %s: %w", id, err))
		return
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
	/*
		// Layer 1: Request validation (validates inline spec fields and metadata)
		if err := validation.ValidateResource(&req); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
			return
		}
	*/
	for _, c := range req.Components {
		// Get version context from request
		// Get version context from request (set by version negotiation middleware)
		versionCtx := versioning.GetVersionContext(r.Context())

		uid, err := resource.GenerateUIDForResource("Component")
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate UID: %w", err))
			return
		}
		// Versioned mode: flat fields with fabrica.Metadata
		component := &v1.Component{
			// Use negotiated ServeVersion (from Accept header) for apiVersion
			APIVersion: versionCtx.GroupVersion,
			Kind:       "Component",
			Spec:       *c,
		}

		component.Metadata.Initialize(c.ID, uid)

		// Layer 2: Fabrica struct tag validation
		if err := validation.ValidateResource(component); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
			return
		}

		// Set labels and annotations
		if component.Metadata.Labels == nil {
			component.Metadata.Labels = make(map[string]string)
		}
		/*
			for k, v := range req.Labels {
				component.Metadata.Labels[k] = v
			}
		*/
		if component.Metadata.Annotations == nil {
			component.Metadata.Annotations = make(map[string]string)
		}
		/*
			for k, v := range req.Annotations {
				component.Metadata.Annotations[k] = v
			}
		*/

		// Layer 2: Custom business logic validation
		if err := validation.ValidateWithContext(r.Context(), component); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
			return
		}

		// Set initial status
		// This assumes the generator passes an 'IsReconcilable' boolean
		// to this template, and that the resource has a .Status.Phase field.

		// Save (Layer 1: Ent validation happens automatically if using Ent storage)
		if err := storage.SaveComponent(r.Context(), component); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save Component: %w", err))
			return
		}
		// Publish resource created event

		if err := events.PublishResourceCreated(r.Context(), "Component", component.Metadata.UID, component.Metadata.Name, component); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource created event for Component %s: %v\n", component.Metadata.UID, err)
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

	component, err := storage.LoadComponentByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to component %s: %w", id, err))
		return
	}

	if component == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("component not found: %s", id))
		return
	}

	var req v1.ComponentSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Apply updates

	// Versioned mode: direct field access
	component.Metadata.Name = req.ID

	// Update spec fields ONLY - status should use /status subresource
	component.Spec = req

	// Update timestamp
	component.Metadata.UpdatedAt = time.Now()

	if err := storage.SaveComponent(r.Context(), component); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save Component: %w", err))
		return
	}

	// Publish resource updated event
	updateMetadata := map[string]interface{}{
		"updatedAt": component.Metadata.UpdatedAt,
	}

	if err := events.PublishResourceUpdated(r.Context(), "Component", component.Metadata.UID, component.Metadata.Name, component, updateMetadata); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource updated event for Component %s: %v\n", component.Metadata.UID, err)

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

	component, err := storage.LoadComponentByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to component %s: %w", id, err))
		return
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

		if err := events.PublishResourceDeleted(r.Context(), "Component", component.Metadata.UID, component.Metadata.Name, deleteMetadata); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource deleted event for Component %s: %v\n", component.Metadata.UID, err)

		}

		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "Component deleted successfully",
			UID:     uid,
		})
	} else {
		// todo maybe do something different
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "Component not present",
			UID:     "",
		})
	}
}
