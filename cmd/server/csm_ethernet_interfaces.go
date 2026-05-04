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

// GetEthernetInterfacesCsm returns all EthernetInterface resources
func GetEthernetInterfacesCsm(w http.ResponseWriter, r *http.Request) {
	ethernetInterfaces, err := plugins.Store.LoadAllEthernetInterfaces(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load ethernetinterfaces: %w", err))
		return
	}

	ethernetInterfaceSpecs := []*v1.EthernetInterfaceSpec{}
	for _, e := range ethernetInterfaces {
		ethernetInterfaceSpecs = append(ethernetInterfaceSpecs, &e.Spec)
	}
	respondJSON(w, http.StatusOK, ethernetInterfaceSpecs)
}

// CreateEthernetInterfaceCsm creates one or more new EthernetInterface resources
func CreateEthernetInterfaceCsm(w http.ResponseWriter, r *http.Request) {
	var req v1.EthernetInterfaceSpec
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
	uid, err := resource.GenerateUIDForResource("EthernetInterface")
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate UID: %w", err))
		return
	}

	ethernetInterface := &v1.EthernetInterface{
		// Use negotiated ServeVersion (from Accept header) for apiVersion
		APIVersion: versionCtx.ServeVersion,
		Kind:       "EthernetInterface",
		Spec:       req,
	}
	// Initialize metadata from request
	ethernetInterface.Metadata.UID = uid
	ethernetInterface.Metadata.Name = req.ID
	now := time.Now()
	ethernetInterface.Metadata.CreatedAt = now
	ethernetInterface.Metadata.UpdatedAt = now

	// Set labels and annotations
	if ethernetInterface.Metadata.Labels == nil {
		ethernetInterface.Metadata.Labels = make(map[string]string)
	}
	if ethernetInterface.Metadata.Annotations == nil {
		ethernetInterface.Metadata.Annotations = make(map[string]string)
	}

	// Layer 2: Custom business logic validation
	if err := validation.ValidateWithContext(r.Context(), ethernetInterface); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
		return
	}

	// Save (Layer 1: Ent validation happens automatically if using Ent storage)
	if err := plugins.Store.SaveEthernetInterface(r.Context(), ethernetInterface); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save EthernetInterface: %w", err))
		return
	}

	// Publish resource created event
	if err := events.PublishResourceCreated(r.Context(), "EthernetInterface", ethernetInterface.Metadata.UID, ethernetInterface.Metadata.Name, ethernetInterface); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource created event for EthernetInterface %s: %v\n", ethernetInterface.Metadata.UID, err)
	}

	respondJSON(w, http.StatusCreated, req)
}

// GetEthernetInterfaceCsm returns a specific EthernetInterface resource by ID
func GetEthernetInterfaceCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("EthernetInterface ID is required"))
		return
	}

	ethernetInterface, err := plugins.Store.LoadEthernetInterfaceByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("failed to find ethernetinterface %s: %w", id, err))
		return
	}

	if ethernetInterface == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("ethernetinterface not found: %s", id))
		return
	}
	respondJSON(w, http.StatusOK, &ethernetInterface.Spec)
}

// UpdateEthernetInterfaceCsm updates the spec of an existing EthernetInterface resource
// NOTE: This endpoint ONLY updates the spec. Use PUT /EthernetInterfaces/{id}/Status to update status.
func UpdateEthernetInterfaceCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("EthernetInterface ID is required"))
		return
	}

	ethernetInterface, err := plugins.Store.LoadEthernetInterfaceByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load ethernetinterface %s: %w", id, err))
		return
	}

	if ethernetInterface == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("ethernetinterface not found: %s", id))
		return
	}

	var req v1.EthernetInterfaceSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	ethernetInterface.Metadata.Name = req.ID
	ethernetInterface.Spec = req

	if ethernetInterface.Metadata.Labels == nil {
		ethernetInterface.Metadata.Labels = make(map[string]string)
	}
	if ethernetInterface.Metadata.Annotations == nil {
		ethernetInterface.Metadata.Annotations = make(map[string]string)
	}

	ethernetInterface.Metadata.UpdatedAt = time.Now()

	if err := plugins.Store.SaveEthernetInterface(r.Context(), ethernetInterface); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save EthernetInterface: %w", err))
		return
	}

	updateMetadata := map[string]interface{}{
		"updatedAt": ethernetInterface.Metadata.UpdatedAt,
	}
	if err := events.PublishResourceUpdated(r.Context(), "EthernetInterface", ethernetInterface.Metadata.UID, ethernetInterface.Metadata.Name, ethernetInterface, updateMetadata); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource updated event for EthernetInterface %s: %v\n", ethernetInterface.Metadata.UID, err)
	}

	respondJSON(w, http.StatusOK, ethernetInterface.Spec)
}

// DeleteEthernetInterfaceCsm deletes an EthernetInterface resource
func DeleteEthernetInterfaceCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("EthernetInterface ID is required"))
		return
	}

	ethernetInterface, err := plugins.Store.LoadEthernetInterfaceByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("failed to find ethernetinterface %s: %w", id, err))
		return
	}

	if ethernetInterface != nil {
		uid := ethernetInterface.GetUID()
		if err := plugins.Store.DeleteEthernetInterface(r.Context(), uid); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete EthernetInterface: %w", err))
			return
		}
		deleteMetadata := map[string]interface{}{
			"deletedAt": time.Now(),
		}
		if err := events.PublishResourceDeleted(r.Context(), "EthernetInterface", ethernetInterface.Metadata.UID, ethernetInterface.Metadata.Name, deleteMetadata); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource deleted event for EthernetInterface %s: %v\n", ethernetInterface.Metadata.UID, err)
		}
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "EthernetInterface deleted successfully",
			UID:     uid,
		})
	} else {
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "EthernetInterface not present",
			UID:     "",
		})
	}
}
