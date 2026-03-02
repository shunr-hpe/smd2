// Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
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

// GetServiceEndpointsSmdV2 returns all ServiceEndpoint resources
func GetServiceEndpointsSmdV2(w http.ResponseWriter, r *http.Request) {
	serviceEndpoints, err := storage.LoadAllServiceEndpoints(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load serviceendpoints: %w", err))
		return
	}
	serviceEndpointArray := ServiceEndpointArray{
		ServiceEndpoints: make([]*v1.ServiceEndpointSpec, len(serviceEndpoints)),
	}
	for i, s := range serviceEndpoints {
		serviceEndpointArray.ServiceEndpoints[i] = &s.Spec
	}
	respondJSON(w, http.StatusOK, serviceEndpointArray)
}

// GetServiceEndpointSmdV2 returns a specific ServiceEndpoint resource by RedfishEndpointID
func GetServiceEndpointSmdV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("ServiceEndpoint RedfishEndpointID is required"))
		return
	}

	serviceEndpoint, err := storage.LoadServiceEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load serviceendpoint %s: %w", id, err))
		return
	}

	if serviceEndpoint == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("serviceendpoint not found: %s", id))
		return
	}
	respondJSON(w, http.StatusOK, &serviceEndpoint.Spec)
}

// CreateServiceEndpointSmdV2 creates one or more new ServiceEndpoint resources
func CreateServiceEndpointSmdV2(w http.ResponseWriter, r *http.Request) {
	var req ServiceEndpointArray
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Get version context from request (set by version negotiation middleware)
	versionCtx := versioning.GetVersionContext(r.Context())
	for _, s := range req.ServiceEndpoints {
		uid, err := resource.GenerateUIDForResource("ServiceEndpoint")
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate UID: %w", err))
			return
		}

		serviceEndpoint := &v1.ServiceEndpoint{
			// Use negotiated ServeVersion (from Accept header) for apiVersion
			APIVersion: versionCtx.ServeVersion,
			Kind:       "ServiceEndpoint",
			Spec:       *s,
		}
		// Initialize metadata from request
		serviceEndpoint.Metadata.UID = uid
		serviceEndpoint.Metadata.Name = s.RfEndpointID
		now := time.Now()
		serviceEndpoint.Metadata.CreatedAt = now
		serviceEndpoint.Metadata.UpdatedAt = now

		// Set labels and annotations
		if serviceEndpoint.Metadata.Labels == nil {
			serviceEndpoint.Metadata.Labels = make(map[string]string)
		}
		if serviceEndpoint.Metadata.Annotations == nil {
			serviceEndpoint.Metadata.Annotations = make(map[string]string)
		}

		// Layer 2: Custom business logic validation
		if err := validation.ValidateWithContext(r.Context(), serviceEndpoint); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
			return
		}

		// Save (Layer 1: Ent validation happens automatically if using Ent storage)
		if err := storage.SaveServiceEndpoint(r.Context(), serviceEndpoint); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save ServiceEndpoint: %w", err))
			return
		}

		// Publish resource created event
		if err := events.PublishResourceCreated(r.Context(), "ServiceEndpoint", serviceEndpoint.Metadata.UID, serviceEndpoint.Metadata.Name, serviceEndpoint); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource created event for ServiceEndpoint %s: %v\n", serviceEndpoint.Metadata.UID, err)
		}
	}

	respondJSON(w, http.StatusCreated, req)
}

// UpdateServiceEndpointSmdV2 updates the spec of an existing ServiceEndpoint resource
// NOTE: This endpoint ONLY updates the spec. Use PUT /ServiceEndpoints/{id}/status to update status.
func UpdateServiceEndpointSmdV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("ServiceEndpoint ID is required"))
		return
	}

	serviceEndpoint, err := storage.LoadServiceEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load serviceendpoint %s: %w", id, err))
		return
	}

	if serviceEndpoint == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("serviceendpoint not found: %s", id))
		return
	}

	var req v1.ServiceEndpointSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	serviceEndpoint.Metadata.Name = req.RfEndpointID
	serviceEndpoint.Spec = req

	if serviceEndpoint.Metadata.Labels == nil {
		serviceEndpoint.Metadata.Labels = make(map[string]string)
	}
	if serviceEndpoint.Metadata.Annotations == nil {
		serviceEndpoint.Metadata.Annotations = make(map[string]string)
	}

	serviceEndpoint.Metadata.UpdatedAt = time.Now()

	if err := storage.SaveServiceEndpoint(r.Context(), serviceEndpoint); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save ServiceEndpoint: %w", err))
		return
	}

	updateMetadata := map[string]interface{}{
		"updatedAt": serviceEndpoint.Metadata.UpdatedAt,
	}
	if err := events.PublishResourceUpdated(r.Context(), "ServiceEndpoint", serviceEndpoint.Metadata.UID, serviceEndpoint.Metadata.Name, serviceEndpoint, updateMetadata); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource updated event for ServiceEndpoint %s: %v\n", serviceEndpoint.Metadata.UID, err)
	}

	respondJSON(w, http.StatusOK, serviceEndpoint.Spec)
}

// DeleteServiceEndpointSmdV2 deletes a ServiceEndpoint resource
func DeleteServiceEndpointSmdV2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("ServiceEndpoint ID is required"))
		return
	}

	serviceEndpoint, err := storage.LoadServiceEndpointByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load serviceendpoint %s: %w", id, err))
		return
	}

	if serviceEndpoint != nil {
		uid := serviceEndpoint.GetUID()
		if err := storage.DeleteServiceEndpoint(r.Context(), uid); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete ServiceEndpoint: %w", err))
			return
		}
		deleteMetadata := map[string]interface{}{
			"deletedAt": time.Now(),
		}
		if err := events.PublishResourceDeleted(r.Context(), "ServiceEndpoint", serviceEndpoint.Metadata.UID, serviceEndpoint.Metadata.Name, deleteMetadata); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource deleted event for ServiceEndpoint %s: %v\n", serviceEndpoint.Metadata.UID, err)
		}
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "ServiceEndpoint deleted successfully",
			UID:     uid,
		})
	} else {
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "ServiceEndpoint not present",
			UID:     "",
		})
	}
}
