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

// GetRedfishEndpointsSmdV2 returns all RedfishEndpoint resources
func GetRedfishEndpointsSmdV2(w http.ResponseWriter, r *http.Request) {
	// Authorization: Add custom middleware in routes.go or implement checks here
	// Example: if !authorized(r) { respondError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized")); return }

	redfishEndpoints, err := storage.LoadAllRedfishEndpoints(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load redfishendpoints: %w", err))
		return
	}
	redfishEndpointArray := RedfishEndpointArray{
		RedfishEndpoints: make([]*v1.RedfishEndpointSpec, len(redfishEndpoints)),
	}
	for i, re := range redfishEndpoints {
		redfishEndpointArray.RedfishEndpoints[i] = &re.Spec
	}
	respondJSON(w, http.StatusOK, redfishEndpointArray)
}

// GetRedfishEndpointSmdV2 returns a specific RedfishEndpoint resource by ID
func GetRedfishEndpointSmdV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("RedfishEndpoint ID is required"))
		return
	}

	// Version context available here for version-aware operations
	// versionCtx := versioning.GetVersionContext(r.Context())
	// Requested version: versionCtx.ServeVersion
	// To enable: replace storage.LoadRedfishEndpoint() with version-aware function

	// Authorization: Add custom middleware in routes.go or implement checks here
	// Example: if !authorized(r) { respondError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized")); return }

	redfishEndpoint, err := storage.LoadRedfishEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load redfishEndpoint %s: %w", id, err))
		return
	}

	if redfishEndpoint == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("redfishEndpoint not found: %s", id))
		return
	}
	respondJSON(w, http.StatusOK, &redfishEndpoint.Spec)
}

// CreateRedfishEndpointSmdV2 creates a new RedfishEndpoint resource
func CreateRedfishEndpointSmdV2(w http.ResponseWriter, r *http.Request) {
	var req v1.RedfishEndpointSpec
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
	uid, err := resource.GenerateUIDForResource("RedfishEndpoint")
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate UID: %w", err))
		return
	}

	// Versioned mode: flat fields with fabrica.Metadata
	redfishEndpoint := &v1.RedfishEndpoint{
		// Use negotiated ServeVersion (from Accept header) for apiVersion
		APIVersion: versionCtx.ServeVersion,
		Kind:       "RedfishEndpoint",
		Spec:       req,
	}
	// Initialize metadata from request
	redfishEndpoint.Metadata.UID = uid
	redfishEndpoint.Metadata.Name = req.ID
	now := time.Now()
	redfishEndpoint.Metadata.CreatedAt = now
	redfishEndpoint.Metadata.UpdatedAt = now

	// Set labels and annotations
	if redfishEndpoint.Metadata.Labels == nil {
		redfishEndpoint.Metadata.Labels = make(map[string]string)
	}
	if redfishEndpoint.Metadata.Annotations == nil {
		redfishEndpoint.Metadata.Annotations = make(map[string]string)
	}

	// Layer 2: Custom business logic validation
	if err := validation.ValidateWithContext(r.Context(), redfishEndpoint); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
		return
	}

	// Save (Layer 1: Ent validation happens automatically if using Ent storage)
	if err := storage.SaveRedfishEndpoint(r.Context(), redfishEndpoint); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save RedfishEndpoint: %w", err))
		return
	}

	// Publish resource created event
	if err := events.PublishResourceCreated(r.Context(), "RedfishEndpoint", redfishEndpoint.Metadata.UID, redfishEndpoint.Metadata.Name, redfishEndpoint); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource created event for RedfishEndpoint %s: %v\n", redfishEndpoint.Metadata.UID, err)
	}

	respondJSON(w, http.StatusCreated, redfishEndpoint.Spec)
}

// UpdateRedfishEndpointSmdV2 updates the spec of an existing RedfishEndpoint resource
// NOTE: This endpoint ONLY updates the spec. Use PUT /redfishendpoints/{id}/status to update status.
func UpdateRedfishEndpointV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("RedfishEndpoint ID is required"))
		return
	}
	redfishEndpoint, err := storage.LoadRedfishEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load redfishEndpoint %s: %w", id, err))
		return
	}

	if redfishEndpoint == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("redfishEndpoint not found: %s", id))
		return
	}

	var req v1.RedfishEndpointSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Apply updates

	// Versioned mode: direct field access
	redfishEndpoint.Metadata.Name = req.ID

	// Update spec fields ONLY - status should use /status subresource
	redfishEndpoint.Spec = req

	// Update labels and annotations
	if redfishEndpoint.Metadata.Labels == nil {
		redfishEndpoint.Metadata.Labels = make(map[string]string)
	}
	if redfishEndpoint.Metadata.Annotations == nil {
		redfishEndpoint.Metadata.Annotations = make(map[string]string)
	}

	// Update timestamp
	redfishEndpoint.Metadata.UpdatedAt = time.Now()

	if err := storage.SaveRedfishEndpoint(r.Context(), redfishEndpoint); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save RedfishEndpoint: %w", err))
		return
	}

	// Publish resource updated event
	updateMetadata := map[string]interface{}{
		"updatedAt": redfishEndpoint.Metadata.UpdatedAt,
	}

	if err := events.PublishResourceUpdated(r.Context(), "RedfishEndpoint", redfishEndpoint.Metadata.UID, redfishEndpoint.Metadata.Name, redfishEndpoint, updateMetadata); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource updated event for RedfishEndpoint %s: %v\n", redfishEndpoint.Metadata.UID, err)
	}

	respondJSON(w, http.StatusOK, redfishEndpoint.Spec)
}

// DeleteRedfishEndpointSmdV2 deletes a RedfishEndpoint resource
func DeleteRedfishEndpointV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("RedfishEndpoint ID is required"))
		return
	}

	redfishEndpoint, err := storage.LoadRedfishEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load redfishEndpoint %s: %w", id, err))
		return
	}

	if redfishEndpoint != nil {
		uid := redfishEndpoint.GetUID()
		if err := storage.DeleteRedfishEndpoint(r.Context(), uid); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete RedfishEndpoint: %w", err))
			return
		}
		// Publish resource deleted event
		deleteMetadata := map[string]interface{}{
			"deletedAt": time.Now(),
		}

		if err := events.PublishResourceDeleted(r.Context(), "RedfishEndpoint", redfishEndpoint.Metadata.UID, redfishEndpoint.Metadata.Name, deleteMetadata); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource deleted event for RedfishEndpoint %s: %v\n", redfishEndpoint.Metadata.UID, err)
		}
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "RedfishEndpoint deleted successfully",
			UID:     uid,
		})
	} else {
		// todo maybe do something different
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "RedfishEndpoint not present",
			UID:     "",
		})
	}
}
