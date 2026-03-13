/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /hsm/v2/State/Components routes registered in csm_routes.go.
 * Routes under test:
 *   GET    /hsm/v2/State/Components
 *   POST   /hsm/v2/State/Components
 *   GET    /hsm/v2/State/Components/{id}
 *   PUT    /hsm/v2/State/Components/{id}
 *   DELETE /hsm/v2/State/Components/{id}
 *
 * Notes on request/response shapes (from csm_component_handlers.go):
 *   POST  body   : ComponentArray  { "Components": [ <ComponentSpec>, ... ] }
 *   POST  returns: HTTP 201, no body
 *   GET / returns: ComponentArray  { "Components": [ <ComponentSpec>, ... ] }
 *   GET /{id}:    full Component   (apiVersion, kind, metadata, spec, status)
 *   PUT  body   : ComponentSpec
 *   PUT  returns: updated ComponentSpec
 *   DELETE /{id}: DeleteResponse { message, uid }
 */

package resttests

import (
	"fmt"
	"net/http"
	"testing"
)

const csmBase = "/hsm/v2/State/Components"

// ─── CSM-specific request / response shapes ───────────────────────────────────
// These mirror csm_models.go / component_types.go without importing package main.

type csmComponentSpec struct {
	ID      string `json:"ID"`
	Type    string `json:"Type,omitempty"`
	State   string `json:"State,omitempty"`
	Flag    string `json:"Flag,omitempty"`
	Role    string `json:"Role,omitempty"`
	SubRole string `json:"SubRole,omitempty"`
	Arch    string `json:"Arch,omitempty"`
	Class   string `json:"Class,omitempty"`
	NID     any    `json:"NID,omitempty"`
}

// csmComponentArray mirrors cmd/server.ComponentArray.
type csmComponentArray struct {
	Components []*csmComponentSpec `json:"Components"`
}

// csmComponentFull mirrors v1.Component returned by GetComponentCsm.
type csmComponentFull struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   componentMetadata `json:"metadata"` // reuse from component_test.go
	Spec       csmComponentSpec  `json:"spec"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// csmCreate POSTs a batch of components via the CSM endpoint and asserts 201.
func csmCreate(t *testing.T, specs ...*csmComponentSpec) {
	t.Helper()
	body := csmComponentArray{Components: specs}
	resp := doRequest(t, http.MethodPost, csmBase, body)
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// csmGetOne fetches a single component by xname ID and returns the full Component.
func csmGetOne(t *testing.T, xname string) (*csmComponentFull, int) {
	t.Helper()
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("%s/%s", csmBase, xname), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var c csmComponentFull
	decodeJSON(t, resp, &c)
	return &c, resp.StatusCode
}

// csmDelete deletes a component by xname ID.
func csmDelete(t *testing.T, xname string) {
	t.Helper()
	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmBase, xname), nil)
	resp.Body.Close()
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateComponentCsm verifies that POST /hsm/v2/State/Components with a
// ComponentArray body returns HTTP 201.
func TestCreateComponentCsm(t *testing.T) {
	xname := "x3000c0s2b0n0"
	csmCreate(t, &csmComponentSpec{ID: xname, Type: "Node"})
	defer csmDelete(t, xname)

	// Verify the component exists after creation
	comp, status := csmGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GET after POST, got %d", status)
	}
	if comp.Spec.ID != xname {
		t.Errorf("expected Spec.ID=%q, got %q", xname, comp.Spec.ID)
	}
}

// TestCreateComponentCsmBulk verifies that multiple components can be created
// in a single POST.
func TestCreateComponentCsmBulk(t *testing.T) {
	xnames := []string{"x3000c0s2b0n1", "x3000c0s2b0n2", "x3000c0s2b0n3"}
	specs := make([]*csmComponentSpec, len(xnames))
	for i, x := range xnames {
		specs[i] = &csmComponentSpec{ID: x, Type: "Node"}
	}
	csmCreate(t, specs...)
	defer func() {
		for _, x := range xnames {
			csmDelete(t, x)
		}
	}()

	// Verify each was created
	for _, x := range xnames {
		comp, status := csmGetOne(t, x)
		if status != http.StatusOK {
			t.Errorf("expected HTTP 200 for %s, got %d", x, status)
			continue
		}
		if comp.Spec.ID != x {
			t.Errorf("expected Spec.ID=%q, got %q", x, comp.Spec.ID)
		}
	}
}

// TestGetComponentsCsm verifies that GET /hsm/v2/State/Components returns
// HTTP 200 and a ComponentArray with at least the created component.
func TestGetComponentsCsm(t *testing.T) {
	xname := "x3000c0s2b0n4"
	csmCreate(t, &csmComponentSpec{ID: xname, Type: "Node"})
	defer csmDelete(t, xname)

	resp := doRequest(t, http.MethodGet, csmBase, nil)
	requireStatus(t, resp, http.StatusOK)

	var list csmComponentArray
	decodeJSON(t, resp, &list)

	found := false
	for _, c := range list.Components {
		if c != nil && c.ID == xname {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("component %s not found in GET %s list", xname, csmBase)
	}
}

// TestGetComponentCsm verifies that GET /hsm/v2/State/Components/{id} returns
// HTTP 200 and the correct component.
func TestGetComponentCsm(t *testing.T) {
	xname := "x3000c0s2b0n5"
	csmCreate(t, &csmComponentSpec{ID: xname, Type: "Node"})
	defer csmDelete(t, xname)

	comp, status := csmGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", status)
	}
	if comp.Spec.ID != xname {
		t.Errorf("expected Spec.ID=%q, got %q", xname, comp.Spec.ID)
	}
	if comp.Kind != "Component" {
		t.Errorf("expected Kind=Component, got %q", comp.Kind)
	}
}

// TestUpdateComponentCsm verifies that PUT /hsm/v2/State/Components/{id}
// updates the component spec and returns HTTP 200.
func TestUpdateComponentCsm(t *testing.T) {
	xname := "x3000c0s2b0n6"
	csmCreate(t, &csmComponentSpec{ID: xname, Type: "Node"})
	defer csmDelete(t, xname)

	updateSpec := csmComponentSpec{
		ID:    xname,
		Type:  "Node",
		State: "Ready",
		Role:  "Compute",
		Flag:  "OK",
	}
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmBase, xname), updateSpec)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Verify via GET that the update persisted
	comp, status := csmGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 after PUT, got %d", status)
	}
	if comp.Spec.State != "Ready" {
		t.Errorf("expected Spec.State=Ready after PUT, got %q", comp.Spec.State)
	}
	if comp.Spec.Role != "Compute" {
		t.Errorf("expected Spec.Role=Compute after PUT, got %q", comp.Spec.Role)
	}
}

// TestDeleteComponentCsm verifies that DELETE /hsm/v2/State/Components/{id}
// returns HTTP 200 and that a subsequent GET does not return HTTP 200.
func TestDeleteComponentCsm(t *testing.T) {
	xname := "x3000c0s2b0n7"
	csmCreate(t, &csmComponentSpec{ID: xname, Type: "Node"})

	// Delete
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmBase, xname), nil)
	requireStatus(t, delResp, http.StatusOK)

	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.Message == "" {
		t.Error("expected non-empty message in delete response")
	}

	// Confirm it is gone – the CSM GET handler returns non-200 for missing components
	_, status := csmGetOne(t, xname)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE %s, got 200", xname)
	}
}

// TestCsmComponentLifecycle exercises the full POST → GET → PUT → DELETE cycle
// via the CSM /hsm/v2/State/Components endpoint.
func TestCsmComponentLifecycle(t *testing.T) {
	xname := "x3000c0s2b0n8"

	// ── POST ──────────────────────────────────────────────────────────────────
	csmCreate(t, &csmComponentSpec{ID: xname, Type: "Node"})

	// ── GET by xname ──────────────────────────────────────────────────────────
	comp, status := csmGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("POST→GET: expected HTTP 200, got %d", status)
	}
	if comp.Spec.ID != xname {
		t.Errorf("POST→GET: expected Spec.ID=%q, got %q", xname, comp.Spec.ID)
	}
	t.Logf("Created component UID: %s", comp.Metadata.UID)

	// ── GET all – component must appear ───────────────────────────────────────
	listResp := doRequest(t, http.MethodGet, csmBase, nil)
	requireStatus(t, listResp, http.StatusOK)
	var list csmComponentArray
	decodeJSON(t, listResp, &list)
	found := false
	for _, c := range list.Components {
		if c != nil && c.ID == xname {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("component %s not found in GET list after creation", xname)
	}

	// ── PUT ───────────────────────────────────────────────────────────────────
	putResp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmBase, xname), csmComponentSpec{
		ID: xname, Type: "Node", State: "On", Role: "Compute", Flag: "OK",
	})
	requireStatus(t, putResp, http.StatusOK)
	putResp.Body.Close()

	comp, status = csmGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("GET after PUT: expected HTTP 200, got %d", status)
	}
	if comp.Spec.State != "On" {
		t.Errorf("PUT: expected Spec.State=On, got %q", comp.Spec.State)
	}

	// ── DELETE ────────────────────────────────────────────────────────────────
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmBase, xname), nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	_, status = csmGetOne(t, xname)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE, still got 200")
	}
}

// TestCreateComponentCsmDuplicateID verifies that POST /hsm/v2/State/Components rejects
// a component whose ID already exists, enforcing resource_id uniqueness.
func TestCreateComponentCsmDuplicateID(t *testing.T) {
	xname := "x3000c0s3b0n0"
	csmCreate(t, &csmComponentSpec{ID: xname, Type: "Node"})
	defer csmDelete(t, xname)

	resp := doRequest(t, http.MethodPost, csmBase, csmComponentArray{
		Components: []*csmComponentSpec{{ID: xname, Type: "Node"}},
	})
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate component ID %q, got HTTP %d", xname, resp.StatusCode)
	}
}
