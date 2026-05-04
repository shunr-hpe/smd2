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

// GetGroupsCsm returns all Group resources
func GetGroupsCsm(w http.ResponseWriter, r *http.Request) {
	groups, err := plugins.Store.LoadAllGroups(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load groups: %w", err))
		return
	}

	groupSpecs := []*v1.GroupSpec{}
	for _, g := range groups {
		groupSpecs = append(groupSpecs, &g.Spec)
	}
	respondJSON(w, http.StatusOK, groupSpecs)
}

// CreateGroupCsm creates a new Group resource
func CreateGroupCsm(w http.ResponseWriter, r *http.Request) {
	var req v1.GroupSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Get version context from request (set by version negotiation middleware)
	versionCtx := versioning.GetVersionContext(r.Context())
	uid, err := resource.GenerateUIDForResource("Group")
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate UID: %w", err))
		return
	}

	group := &v1.Group{
		// Use negotiated ServeVersion (from Accept header) for apiVersion
		APIVersion: versionCtx.ServeVersion,
		Kind:       "Group",
		Spec:       req,
	}
	// Initialize metadata from request
	group.Metadata.UID = uid
	group.Metadata.Name = req.Label
	now := time.Now()
	group.Metadata.CreatedAt = now
	group.Metadata.UpdatedAt = now

	// Set labels and annotations
	if group.Metadata.Labels == nil {
		group.Metadata.Labels = make(map[string]string)
	}
	if group.Metadata.Annotations == nil {
		group.Metadata.Annotations = make(map[string]string)
	}

	// Layer 2: Custom business logic validation
	if err := validation.ValidateWithContext(r.Context(), group); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("validation failed: %w", err))
		return
	}

	// Save (Layer 1: Ent validation happens automatically if using Ent storage)
	if err := plugins.Store.SaveGroup(r.Context(), group); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save Group: %w", err))
		return
	}

	// Publish resource created event
	if err := events.PublishResourceCreated(r.Context(), "Group", group.Metadata.UID, group.Metadata.Name, group); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource created event for Group %s: %v\n", group.Metadata.UID, err)
	}

	respondJSON(w, http.StatusCreated, req)
}

// GetGroupCsm returns a specific Group resource by label
func GetGroupCsm(w http.ResponseWriter, r *http.Request) {
	label := chi.URLParam(r, "group_label")
	if label == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Group label is required"))
		return
	}

	group, err := plugins.Store.LoadGroupByLabel(r.Context(), label)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("failed to find group %s: %w", label, err))
		return
	}

	if group == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("group not found: %s", label))
		return
	}
	respondJSON(w, http.StatusOK, &group.Spec)
}

// UpdateGroupCsm updates the spec of an existing Group resource
// NOTE: This endpoint ONLY updates the spec. Use PUT /groups/{group_label}/status to update status.
func UpdateGroupCsm(w http.ResponseWriter, r *http.Request) {
	label := chi.URLParam(r, "group_label")
	if label == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Group label is required"))
		return
	}

	group, err := plugins.Store.LoadGroupByLabel(r.Context(), label)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load group %s: %w", label, err))
		return
	}

	if group == nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("group not found: %s", label))
		return
	}

	var req v1.GroupSpec
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	group.Metadata.Name = req.Label
	group.Spec = req

	if group.Metadata.Labels == nil {
		group.Metadata.Labels = make(map[string]string)
	}
	if group.Metadata.Annotations == nil {
		group.Metadata.Annotations = make(map[string]string)
	}

	group.Metadata.UpdatedAt = time.Now()

	if err := plugins.Store.SaveGroup(r.Context(), group); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to save Group: %w", err))
		return
	}

	updateMetadata := map[string]interface{}{
		"updatedAt": group.Metadata.UpdatedAt,
	}
	if err := events.PublishResourceUpdated(r.Context(), "Group", group.Metadata.UID, group.Metadata.Name, group, updateMetadata); err != nil {
		// Log the error but don't fail the request - events are non-critical
		fmt.Printf("Warning: Failed to publish resource updated event for Group %s: %v\n", group.Metadata.UID, err)
	}

	respondJSON(w, http.StatusOK, group.Spec)
}

// DeleteGroupCsm deletes a Group resource
func DeleteGroupCsm(w http.ResponseWriter, r *http.Request) {
	label := chi.URLParam(r, "group_label")
	if label == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("Group label is required"))
		return
	}

	group, err := plugins.Store.LoadGroupByLabel(r.Context(), label)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Errorf("failed to find group %s: %w", label, err))
		return
	}

	if group != nil {
		uid := group.GetUID()
		if err := plugins.Store.DeleteGroup(r.Context(), uid); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete Group: %w", err))
			return
		}
		deleteMetadata := map[string]interface{}{
			"deletedAt": time.Now(),
		}
		if err := events.PublishResourceDeleted(r.Context(), "Group", group.Metadata.UID, group.Metadata.Name, deleteMetadata); err != nil {
			// Log the error but don't fail the request - events are non-critical
			fmt.Printf("Warning: Failed to publish resource deleted event for Group %s: %v\n", group.Metadata.UID, err)
		}
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "Group deleted successfully",
			UID:     uid,
		})
	} else {
		respondJSON(w, http.StatusOK, &DeleteResponse{
			Message: "Group not present",
			UID:     "",
		})
	}
}
