// Copyright © 2025-2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

// CreateRedfishEndpointSmdV2 creates a new RedfishEndpoint resource.
// Accepts two input formats:
//  1. Standard format: a single v1.RedfishEndpointSpec JSON object.
//  2. V2 inventory format (from OpenCHAMI/smd parseRedfishEndpointDataV2): a JSON object
//     with the same endpoint-level fields (ID, FQDN, …) plus optional "Systems" and
//     "Managers" inventory arrays. The presence of either array indicates V2 format.
func CreateRedfishEndpointSmdV2(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("failed to read request body: %w", err))
		return
	}

	// Always unmarshal as V2Request — if Systems/Managers are absent it behaves identically
	// to the plain RedfishEndpointSpec because RedfishEndpointSpec is embedded.
	var v2req RedfishEndpointV2Request
	if err := json.Unmarshal(body, &v2req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	isV2Format := len(v2req.Systems) > 0 || len(v2req.Managers) > 0
	if isV2Format {
		fmt.Printf("Info: CreateRedfishEndpointSmdV2: received V2 inventory format for endpoint %s (systems=%d, managers=%d)\n",
			v2req.ID, len(v2req.Systems), len(v2req.Managers))
	}

	req := v2req.RedfishEndpointSpec

	// Apply defaults for fields that are not set in the request.
	if req.FQDN == "" {
		req.FQDN = req.ID
	}
	if !req.Enabled {
		req.Enabled = true
	}
	if req.DiscoveryInfo.LastStatus == "" {
		req.DiscoveryInfo.LastStatus = "NotYetQueried"
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

	// If V2 format: create Components, ComponentEndpoints, and EthernetInterfaces
	// for every entry in the Systems and Managers arrays, mirroring what
	// parseRedfishEndpointDataV2 does in OpenCHAMI/smd.
	if isV2Format {
		if err := createV2SubResources(r.Context(), req, v2req, versionCtx); err != nil {
			// Log but do not fail — the RedfishEndpoint itself was saved successfully.
			fmt.Printf("Warning: CreateRedfishEndpointSmdV2: failed to create V2 sub-resources for %s: %v\n", req.ID, err)
		}
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

// createV2SubResources processes the Systems and Managers arrays in a V2-format POST body,
// creating Component, ComponentEndpoint, and EthernetInterface resources for each entry.
// This mirrors the behaviour of parseRedfishEndpointDataV2 in OpenCHAMI/smd.
//
// For each Manager:
//   - Creates a NodeBMC Component (ID = endpoint.ID)
//   - Creates a Manager ComponentEndpoint (ComponentEndpointType "ComponentEndpointBMC")
//   - Creates an EthernetInterface for every manager ethernet interface that has a MAC address
//
// For each System (index i → suffix "n<i>"):
//   - Creates a Node Component (ID = endpoint.ID + "n<i>")
//   - Creates a ComputerSystem ComponentEndpoint (ComponentEndpointType "ComponentEndpointComputerSystem")
//   - Creates an EthernetInterface for every system ethernet interface that has a MAC address
func createV2SubResources(
	ctx context.Context,
	endpoint v1.RedfishEndpointSpec,
	v2req RedfishEndpointV2Request,
	versionCtx *versioning.VersionContext,
) error {
	now := time.Now()

	// fqdn is the endpoint FQDN, falling back to endpoint ID when FQDN is empty.
	fqdn := endpoint.FQDN
	if fqdn == "" {
		fqdn = endpoint.ID
	}

	// extractPath strips scheme and host from a URI, returning just the path.
	// If the URI is already a path (no host), it is returned unchanged.
	extractPath := func(uri string) string {
		if u, err := url.Parse(uri); err == nil && u.Host != "" {
			return u.Path
		}
		return uri
	}

	// Build []*v1.EthernetNICInfo from V2 ethernet interface entries.
	// InterfaceEnabled defaults to true; the raw eth.Enabled field is omitempty
	// and therefore false when absent from the input payload.
	nicInfoFromEths := func(eths []RedfishEndpointV2EthernetInterface) []*v1.EthernetNICInfo {
		out := make([]*v1.EthernetNICInfo, 0, len(eths))
		for _, eth := range eths {
			eth := eth // capture
			enabled := true
			out = append(out, &v1.EthernetNICInfo{
				InterfaceEnabled: &enabled,
				RedfishId:        eth.URI,
				Oid:              eth.URI,
				Description:      eth.Description,
				MACAddress:       eth.MAC,
			})
		}
		return out
	}

	// saveEthInterfaces creates EthernetInterface resources from V2 ethernet entries.
	saveEthInterfaces := func(compID, compType string, eths []RedfishEndpointV2EthernetInterface) error {
		for _, eth := range eths {
			if eth.MAC == "" {
				continue
			}
			macID := strings.ReplaceAll(strings.ToLower(eth.MAC), ":", "")
			uid, err := resource.GenerateUIDForResource("EthernetInterface")
			if err != nil {
				return fmt.Errorf("failed to generate UID for EthernetInterface %s: %w", macID, err)
			}
			var ipAddresses []v1.IPAddress
			if eth.IP != "" {
				ipAddresses = []v1.IPAddress{{IPAddress: eth.IP}}
			}
			ei := &v1.EthernetInterface{
				APIVersion: versionCtx.ServeVersion,
				Kind:       "EthernetInterface",
				Spec: v1.EthernetInterfaceSpec{
					ID:          macID,
					Description: eth.Description,
					MACAddr:     eth.MAC,
					IPAddresses: ipAddresses,
					LastUpdate:  now.UTC().Format(time.RFC3339Nano),
					CompID:      compID,
					Type:        compType,
				},
			}
			ei.Metadata.UID = uid
			ei.Metadata.Name = macID
			ei.Metadata.CreatedAt = now
			ei.Metadata.UpdatedAt = now
			ei.Metadata.Labels = make(map[string]string)
			ei.Metadata.Annotations = make(map[string]string)
			if err := storage.SaveEthernetInterface(ctx, ei); err != nil {
				return fmt.Errorf("failed to save EthernetInterface %s: %w", macID, err)
			}
		}
		return nil
	}

	// ── Managers → NodeBMC Component + Manager ComponentEndpoint + EthernetInterfaces ──
	for _, manager := range v2req.Managers {
		compID := endpoint.ID

		// Component (NodeBMC)
		compUID, err := resource.GenerateUIDForResource("Component")
		if err != nil {
			return fmt.Errorf("failed to generate UID for NodeBMC Component %s: %w", compID, err)
		}
		enabled := true
		comp := &v1.Component{
			APIVersion: versionCtx.ServeVersion,
			Kind:       "Component",
			Spec: v1.ComponentSpec{
				ID:      compID,
				Type:    "NodeBMC",
				Enabled: &enabled,
			},
		}
		comp.Metadata.UID = compUID
		comp.Metadata.Name = compID
		comp.Metadata.CreatedAt = now
		comp.Metadata.UpdatedAt = now
		comp.Metadata.Labels = make(map[string]string)
		comp.Metadata.Annotations = make(map[string]string)
		if err := storage.SaveComponent(ctx, comp); err != nil {
			return fmt.Errorf("failed to save NodeBMC Component %s: %w", compID, err)
		}

		// EthernetInterfaces for this manager
		if err := saveEthInterfaces(compID, "NodeBMC", manager.EthernetInterfaces); err != nil {
			return err
		}
	}

	// ── Systems → Node Component + ComputerSystem ComponentEndpoint + EthernetInterfaces ──
	for i, system := range v2req.Systems {
		nodeID := fmt.Sprintf("%sn%d", endpoint.ID, i)

		enabled := true
		comp, err := storage.LoadComponentByID(ctx, nodeID)
		if err == storage.ErrNotFound {
			// Component (Node)
			compUID, err := resource.GenerateUIDForResource("Component")
			if err != nil {
				return fmt.Errorf("failed to generate UID for Node Component %s: %w", nodeID, err)
			}
			comp := &v1.Component{
				APIVersion: versionCtx.ServeVersion,
				Kind:       "Component",
				Spec: v1.ComponentSpec{
					ID:      nodeID,
					Type:    "Node",
					State:   "On",
					Enabled: &enabled,
					Role:    "Compute",
				},
			}
			comp.Metadata.UID = compUID
			comp.Metadata.Name = nodeID
			comp.Metadata.CreatedAt = now
			comp.Metadata.UpdatedAt = now
			comp.Metadata.Labels = make(map[string]string)
			comp.Metadata.Annotations = make(map[string]string)

		} else if err != nil {
			// unexpected storage error
			return fmt.Errorf("Failed to load component %s: %w", nodeID, err)
		} else {
			comp.Spec.Type = "Node"
			comp.Spec.State = "On"
			comp.Spec.Enabled = &enabled
			comp.Spec.Role = "Compute"

			comp.Metadata.UpdatedAt = now
		}

		if err := storage.SaveComponent(ctx, comp); err != nil {
			return fmt.Errorf("failed to save Node Component %s: %w", nodeID, err)
		}

		// ComponentEndpoint (ComputerSystem)
		cepUID, err := resource.GenerateUIDForResource("ComponentEndpoint")
		if err != nil {
			return fmt.Errorf("failed to generate UID for System ComponentEndpoint %s: %w", nodeID, err)
		}
		systemPath := extractPath(system.URI)
		cep := &v1.ComponentEndpoint{
			APIVersion: versionCtx.ServeVersion,
			Kind:       "ComponentEndpoint",
			Spec: v1.ComponentEndpointSpec{
				ID:                    nodeID,
				Type:                  "Node",
				RedfishType:           "ComputerSystem",
				RedfishSubtype:        system.SystemType,
				UUID:                  system.UUID,
				OdataID:               systemPath,
				RfEndpointID:          endpoint.ID,
				RedfishEndpointFQDN:   fqdn,
				URL:                   fqdn + systemPath,
				ComponentEndpointType: "ComponentEndpointComputerSystem",
				Enabled:               true,
				RedfishSystemInfo: &v1.ComponentSystemInfo{
					Name: system.Name,
					Actions: &v1.ComputerSystemActions{
						ComputerSystemReset: v1.ActionReset{
							AllowableValues: []string{
								"On", "ForceOff", "GracefulShutdown", "GracefulRestart",
								"ForceRestart", "Nmi", "ForceOn", "PushPowerButton",
								"PowerCycle", "Suspend", "Pause", "Resume",
							},
							RFActionInfo: systemPath + "/ResetActionInfo",
							Target:       systemPath + "/Actions/ComputerSystem.Reset",
						},
					},
					EthNICInfo: nicInfoFromEths(system.EthernetInterfaces),
				},
			},
		}
		cep.Metadata.UID = cepUID
		cep.Metadata.Name = nodeID
		cep.Metadata.CreatedAt = now
		cep.Metadata.UpdatedAt = now
		cep.Metadata.Labels = make(map[string]string)
		cep.Metadata.Annotations = make(map[string]string)
		if err := storage.SaveComponentEndpoint(ctx, cep); err != nil {
			return fmt.Errorf("failed to save System ComponentEndpoint %s: %w", nodeID, err)
		}

		// EthernetInterfaces for this system
		if err := saveEthInterfaces(nodeID, "Node", system.EthernetInterfaces); err != nil {
			return err
		}
	}

	return nil
}
