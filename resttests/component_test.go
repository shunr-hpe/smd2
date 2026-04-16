/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /components routes registered in routes_generated.go.
 * Routes under test:
 *   GET    /components
 *   POST   /components
 *   GET    /components/{uid}
 *   PUT    /components/{uid}
 *   DELETE /components/{uid}
 */

package resttests

import (
	"encoding/json"
	"net/http"
	"testing"
)

// ─── Request / response shapes ────────────────────────────────────────────────
// These mirror the structures in cmd/server/models_generated.go and
// apis/inventory-service.openchami.org/v1/component_types.go without importing package main.

type componentMetadata struct {
	Name      string            `json:"name,omitempty"`
	UID       string            `json:"uid,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt string            `json:"createdAt,omitempty"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
}

type componentSpec struct {
	ID      string `json:"ID"`
	Type    string `json:"Type,omitempty"`
	State   string `json:"State,omitempty"`
	Flag    string `json:"Flag,omitempty"`
	Role    string `json:"Role,omitempty"`
	SubRole string `json:"SubRole,omitempty"`
	Arch    string `json:"Arch,omitempty"`
	Class   string `json:"Class,omitempty"`
	NID     any    `json:"NID,omitempty"`
	NetType string `json:"NetType,omitempty"`
	Enabled *bool  `json:"Enabled,omitempty"`
}

type componentResponse struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   componentMetadata `json:"metadata"`
	Spec       componentSpec     `json:"spec"`
}

type createComponentRequest struct {
	Metadata componentMetadata `json:"metadata"`
	Spec     componentSpec     `json:"spec"`
}

type updateComponentRequest struct {
	Metadata componentMetadata `json:"metadata,omitempty"`
	Spec     componentSpec     `json:"spec,omitempty"`
}

type deleteComponentResponse struct {
	Message string `json:"message"`
	UID     string `json:"uid"`
}

// newComponent is a helper that builds a valid create request.
func newComponent(xname, nodeType string) createComponentRequest {
	enabled := true
	return createComponentRequest{
		Metadata: componentMetadata{Name: xname},
		Spec: componentSpec{
			ID:      xname,
			Type:    nodeType,
			State:   "On",
			Flag:    "OK",
			Role:    "Compute",
			NID:     3,
			Arch:    "X86",
			Class:   "River",
			Enabled: &enabled,
			NetType: "Sling",
		},
	}
}

// createAndRequire POSTs a component and fails the test if the response is not 201.
// It returns the parsed component from the response body.
func createAndRequire(t *testing.T, req createComponentRequest) componentResponse {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/components", req)
	defer resp.Body.Close()
	requireStatus(t, resp, http.StatusCreated)
	var created componentResponse
	decodeJSON(t, resp, &created)
	if created.Metadata.UID == "" {
		t.Fatal("expected non-empty UID in created component response")
	}
	return created
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateComponent verifies that POST /components returns HTTP 201 and a
// Component resource with a generated UID.
func TestCreateComponent(t *testing.T) {
	req := newComponent("x3000c0s0b0n0", "Node")
	component := createAndRequire(t, req)

	if component.Kind != "Component" {
		t.Errorf("expected Kind=Component, got %q", component.Kind)
	}
	if component.Spec.ID != req.Spec.ID {
		t.Errorf("expected Spec.ID=%q, got %q", req.Spec.ID, component.Spec.ID)
	}
	if component.Spec.Type != req.Spec.Type {
		t.Errorf("expected Spec.Type=%q, got %q", req.Spec.Type, component.Spec.Type)
	}

	// Clean up
	doRequest(t, http.MethodDelete, "/components/"+component.Metadata.UID, nil).Body.Close()
}

// TestGetComponents verifies that GET /components returns HTTP 200 and a JSON
// array (which may be empty or populated).
func TestGetComponents(t *testing.T) {
	// Create a component so the list is not empty
	created := createAndRequire(t, newComponent("x3000c0s0b0n1", "Node"))
	defer func() { doRequest(t, http.MethodDelete, "/components/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/components", nil)
	requireStatus(t, resp, http.StatusOK)

	var components []componentResponse
	decodeJSON(t, resp, &components)

	if len(components) == 0 {
		t.Error("expected at least one component in list, got zero")
	}
}

// TestGetComponent verifies that GET /components/{uid} returns HTTP 200 and
// the correct component.
func TestGetComponent(t *testing.T) {
	created := createAndRequire(t, newComponent("x3000c0s0b0n2", "Node"))
	defer func() { doRequest(t, http.MethodDelete, "/components/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/components/"+created.Metadata.UID, nil)
	requireStatus(t, resp, http.StatusOK)

	var fetched componentResponse
	decodeJSON(t, resp, &fetched)

	if fetched.Metadata.UID != created.Metadata.UID {
		t.Errorf("expected UID=%q, got %q", created.Metadata.UID, fetched.Metadata.UID)
	}
	if fetched.Spec.ID != created.Spec.ID {
		t.Errorf("expected Spec.ID=%q, got %q", created.Spec.ID, fetched.Spec.ID)
	}
}

// TestGetComponentNotFound verifies that GET /components/{uid} for an unknown
// UID returns HTTP 404.
func TestGetComponentNotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/components/component-does-not-exist", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 for unknown component, got %d", resp.StatusCode)
	}
}

// TestUpdateComponent verifies that PUT /components/{uid} returns HTTP 200 and
// that the updated fields are persisted.
func TestUpdateComponent(t *testing.T) {
	created := createAndRequire(t, newComponent("x3000c0s0b0n3", "Node"))
	defer func() { doRequest(t, http.MethodDelete, "/components/"+created.Metadata.UID, nil).Body.Close() }()

	updateReq := updateComponentRequest{
		Metadata: componentMetadata{Name: "x3000c0s0b0n3"},
		Spec: componentSpec{
			ID:    "x3000c0s0b0n3",
			Type:  "Node",
			State: "Ready",
			Role:  "Compute",
		},
	}

	resp := doRequest(t, http.MethodPut, "/components/"+created.Metadata.UID, updateReq)
	requireStatus(t, resp, http.StatusOK)

	var updated componentResponse
	decodeJSON(t, resp, &updated)

	if updated.Spec.State != "Ready" {
		t.Errorf("expected Spec.State=Ready after update, got %q", updated.Spec.State)
	}
	if updated.Spec.Role != "Compute" {
		t.Errorf("expected Spec.Role=Compute after update, got %q", updated.Spec.Role)
	}

	// Confirm the change persisted via a subsequent GET
	getResp := doRequest(t, http.MethodGet, "/components/"+created.Metadata.UID, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched componentResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.State != "Ready" {
		t.Errorf("GET after PUT: expected Spec.State=Ready, got %q", fetched.Spec.State)
	}
}

// TestDeleteComponent verifies that DELETE /components/{uid} returns HTTP 200
// and that a subsequent GET returns HTTP 404.
func TestDeleteComponent(t *testing.T) {
	created := createAndRequire(t, newComponent("x3000c0s0b0n4", "Node"))

	// Delete
	delResp := doRequest(t, http.MethodDelete, "/components/"+created.Metadata.UID, nil)
	requireStatus(t, delResp, http.StatusOK)

	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.UID != created.Metadata.UID {
		t.Errorf("delete response UID mismatch: want %q, got %q", created.Metadata.UID, delBody.UID)
	}

	// Verify it is gone
	getResp := doRequest(t, http.MethodGet, "/components/"+created.Metadata.UID, nil)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", getResp.StatusCode)
	}
}

// TestComponentLifecycle exercises the full POST → GET → PUT → DELETE cycle
// in one test so the end-to-end flow is clearly visible.
func TestComponentLifecycle(t *testing.T) {
	xname := "x3000c0s0b0n5"

	// ── POST ──────────────────────────────────────────────────────────────────
	created := createAndRequire(t, newComponent(xname, "Node"))
	uid := created.Metadata.UID
	t.Logf("Created component UID: %s", uid)

	// ── GET by UID ────────────────────────────────────────────────────────────
	getResp := doRequest(t, http.MethodGet, "/components/"+uid, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched componentResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.ID != xname {
		t.Errorf("GET: expected Spec.ID=%q, got %q", xname, fetched.Spec.ID)
	}

	// ── GET list – component must appear ─────────────────────────────────────
	listResp := doRequest(t, http.MethodGet, "/components", nil)
	requireStatus(t, listResp, http.StatusOK)
	var list []json.RawMessage
	decodeJSON(t, listResp, &list)
	found := false
	for _, raw := range list {
		var c componentResponse
		if err := json.Unmarshal(raw, &c); err == nil && c.Metadata.UID == uid {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("component %s not found in GET /components list", uid)
	}

	// ── PUT ───────────────────────────────────────────────────────────────────
	putResp := doRequest(t, http.MethodPut, "/components/"+uid, updateComponentRequest{
		Metadata: componentMetadata{Name: xname},
		Spec:     componentSpec{ID: xname, Type: "Node", State: "On", Flag: "OK", Role: "Compute"},
	})
	requireStatus(t, putResp, http.StatusOK)
	var updated componentResponse
	decodeJSON(t, putResp, &updated)
	if updated.Spec.State != "On" {
		t.Errorf("PUT: expected Spec.State=On, got %q", updated.Spec.State)
	}

	// ── DELETE ────────────────────────────────────────────────────────────────
	delResp := doRequest(t, http.MethodDelete, "/components/"+uid, nil)
	requireStatus(t, delResp, http.StatusOK)

	// Confirm deletion
	gone := doRequest(t, http.MethodGet, "/components/"+uid, nil)
	defer gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", gone.StatusCode)
	}
}

// TestCreateComponentDuplicateID verifies that POST /components rejects a second
// resource with the same Spec.ID, enforcing resource_id uniqueness.
func TestCreateComponentDuplicateID(t *testing.T) {
	xname := "x3001c0s0b0n0"
	first := createAndRequire(t, newComponent(xname, "Node"))
	defer func() { doRequest(t, http.MethodDelete, "/components/"+first.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodPost, "/components", newComponent(xname, "Node"))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate component ID %q, got HTTP %d", xname, resp.StatusCode)
	}
}
