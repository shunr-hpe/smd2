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

// GetHardwaresCsm returns all Hardware resources
func GetHardwaresCsm(w http.ResponseWriter, r *http.Request) {
	hardwares, err := plugins.Store.LoadAllHardwares(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load hardwares: %w", err))
		return
	}
	hardwareCsm := make([]*v1.HardwareSpec, len(hardwares))
	for i, h := range hardwares {
		hardwareCsm[i] = &h.Spec
	}
	respondJSON(w, http.StatusOK, hardwareCsm)
}

// GetHardwareCsm returns a specific Hardware resource by xname (Spec.ID)
func GetHardwareCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Hardware ID is required"))
		return
	}

	hardware, err := plugins.Store.LoadHardwareByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("failed to load hardware %s: %w", id, err))
		return
	}

	if hardware == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("hardware not found: %s", id))
		return
	}
	respondJSON(w, http.StatusOK, &hardware.Spec)
}

// CreateHardwareCsm creates one or more Hardware resources
func CreateHardwareCsm(w http.ResponseWriter, r *http.Request) {
	var req HardwareArray
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	versionCtx := versioning.GetVersionContext(r.Context())
	for _, h := range req.Hardware {
		uid, err := resource.GenerateUIDForResource("Hardware")
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate UID: %w", err))
			return
		}

		hardware := &v1.Hardware{
			APIVersion: versionCtx.ServeVersion,
			Kind:       "Hardware",
			Spec:       *h,
		}
		hardware.Metadata.UID = uid
		hardware.Metadata.Name = h.ID
		now := time.Now()
		hardware.Metadata.CreatedAt = now
		hardware.Metadata.UpdatedAt = now

		if hardware.Metadata.Labels == nil {
			hardware.Metadata.Labels = make(map[string]string)
		}
		if hardware.Metadata.Annotations == nil {
			hardware.Metadata.Annotations = make(map[string]string)
		}

		if err := validation.ValidateWithContext(r.Context(), hardware); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
			return
		}

		if err := plugins.Store.SaveHardware(r.Context(), hardware); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save Hardware: %w", err))
			return
		}

		if err := events.PublishResourceCreated(r.Context(), "Hardware", hardware.Metadata.UID, hardware.Metadata.Name, hardware); err != nil {
			fmt.Printf("Warning: Failed to publish resource created event for Hardware %s: %v\n", hardware.Metadata.UID, err)
		}
	}

	respondJSON(w, http.StatusCreated, req)
}

// UpdateHardwareCsm updates an existing Hardware resource by xname (Spec.ID)
func UpdateHardwareCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Hardware ID is required"))
		return
	}

	hardware, err := plugins.Store.LoadHardwareByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load hardware %s: %w", id, err))
		return
	}
	if hardware == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("hardware not found: %s", id))
		return
	}

	var req v1.HardwareSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	hardware.Metadata.Name = req.ID
	hardware.Spec = req

	if hardware.Metadata.Labels == nil {
		hardware.Metadata.Labels = make(map[string]string)
	}
	if hardware.Metadata.Annotations == nil {
		hardware.Metadata.Annotations = make(map[string]string)
	}

	hardware.Metadata.UpdatedAt = time.Now()

	if err := plugins.Store.SaveHardware(r.Context(), hardware); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save Hardware: %w", err))
		return
	}

	updateMetadata := map[string]interface{}{
		"updatedAt": hardware.Metadata.UpdatedAt,
	}
	if err := events.PublishResourceUpdated(r.Context(), "Hardware", hardware.Metadata.UID, hardware.Metadata.Name, hardware, updateMetadata); err != nil {
		fmt.Printf("Warning: Failed to publish resource updated event for Hardware %s: %v\n", hardware.Metadata.UID, err)
	}

	respondJSON(w, http.StatusOK, hardware.Spec)
}

// DeleteHardwareCsm deletes a Hardware resource by xname (Spec.ID)
func DeleteHardwareCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Hardware ID is required"))
		return
	}

	hardware, err := plugins.Store.LoadHardwareByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("failed to find hardware %s: %w", id, err))
		return
	}

	if hardware != nil {
		uid := hardware.GetUID()
		if err := plugins.Store.DeleteHardware(r.Context(), uid); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete Hardware: %w", err))
			return
		}

		deleteMetadata := map[string]interface{}{
			"deletedAt": time.Now(),
		}
		if err := events.PublishResourceDeleted(r.Context(), "Hardware", hardware.Metadata.UID, hardware.Metadata.Name, deleteMetadata); err != nil {
			fmt.Printf("Warning: Failed to publish resource deleted event for Hardware %s: %v\n", hardware.Metadata.UID, err)
		}

		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "Hardware deleted successfully",
			UID:     uid,
		})
	} else {
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "Hardware not present",
			UID:     "",
		})
	}
}
