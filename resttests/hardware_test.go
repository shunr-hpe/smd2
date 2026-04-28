/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /hardwares routes registered in routes_generated.go.
 * Routes under test:
 *   GET    /hardwares
 *   POST   /hardwares
 *   GET    /hardwares/{uid}
 *   PUT    /hardwares/{uid}
 *   DELETE /hardwares/{uid}
 */

package resttests

import (
	"encoding/json"
	"net/http"
	"testing"
)

// ─── Request / response shapes ────────────────────────────────────────────────
// These mirror the structures in cmd/server/models_generated.go and
// apis/inventory-service.openchami.org/v1/hardware_types.go without importing package main.

type hardwareSpec struct {
	ID                        string `json:"ID"`
	Type                      string `json:"Type,omitempty"`
	Ordinal                   int    `json:"Ordinal,omitempty"`
	Status                    string `json:"Status,omitempty"`
	HWInventoryByLocationType string `json:"HWInventoryByLocationType,omitempty"`
}

type hardwareResponse struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   componentMetadata `json:"metadata"`
	Spec       hardwareSpec      `json:"spec"`
}

type createHardwareRequest struct {
	Metadata componentMetadata `json:"metadata"`
	Spec     hardwareSpec      `json:"spec"`
}

type updateHardwareRequest struct {
	Metadata componentMetadata `json:"metadata,omitempty"`
	Spec     hardwareSpec      `json:"spec,omitempty"`
}

// newHardware builds a valid create request for a hardware item.
func newHardware(xname, hwType string) createHardwareRequest {
	return createHardwareRequest{
		Metadata: componentMetadata{Name: xname},
		Spec: hardwareSpec{
			ID:     xname,
			Type:   hwType,
			Status: "Populated",
		},
	}
}

// createHardwareAndRequire POSTs a hardware item and fails the test if the response is not 201.
// It returns the parsed hardware from the response body.
func createHardwareAndRequire(t *testing.T, req createHardwareRequest) hardwareResponse {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/hardwares", req)
	defer resp.Body.Close()
	requireStatus(t, resp, http.StatusCreated)
	var created hardwareResponse
	decodeJSON(t, resp, &created)
	if created.Metadata.UID == "" {
		t.Fatal("expected non-empty UID in created hardware response")
	}
	return created
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateHardware verifies that POST /hardwares returns HTTP 201 and a
// Hardware resource with a generated UID.
func TestCreateHardware(t *testing.T) {
	req := newHardware("x3000c0s5b0n0", "Node")
	hw := createHardwareAndRequire(t, req)

	if hw.Kind != "Hardware" {
		t.Errorf("expected Kind=Hardware, got %q", hw.Kind)
	}
	if hw.Spec.ID != req.Spec.ID {
		t.Errorf("expected Spec.ID=%q, got %q", req.Spec.ID, hw.Spec.ID)
	}
	if hw.Spec.Type != req.Spec.Type {
		t.Errorf("expected Spec.Type=%q, got %q", req.Spec.Type, hw.Spec.Type)
	}

	// Clean up
	doRequest(t, http.MethodDelete, "/hardwares/"+hw.Metadata.UID, nil).Body.Close()
}

// TestGetHardwares verifies that GET /hardwares returns HTTP 200 and a JSON
// array with at least the created hardware item.
func TestGetHardwares(t *testing.T) {
	created := createHardwareAndRequire(t, newHardware("x3000c0s5b0n1", "Node"))
	defer func() { doRequest(t, http.MethodDelete, "/hardwares/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/hardwares", nil)
	requireStatus(t, resp, http.StatusOK)

	var hardwares []hardwareResponse
	decodeJSON(t, resp, &hardwares)

	if len(hardwares) == 0 {
		t.Error("expected at least one hardware in list, got zero")
	}
}

// TestGetHardware verifies that GET /hardwares/{uid} returns HTTP 200 and
// the correct hardware item.
func TestGetHardware(t *testing.T) {
	created := createHardwareAndRequire(t, newHardware("x3000c0s5b0n2", "Node"))
	defer func() { doRequest(t, http.MethodDelete, "/hardwares/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/hardwares/"+created.Metadata.UID, nil)
	requireStatus(t, resp, http.StatusOK)

	var fetched hardwareResponse
	decodeJSON(t, resp, &fetched)

	if fetched.Metadata.UID != created.Metadata.UID {
		t.Errorf("expected UID=%q, got %q", created.Metadata.UID, fetched.Metadata.UID)
	}
	if fetched.Spec.ID != created.Spec.ID {
		t.Errorf("expected Spec.ID=%q, got %q", created.Spec.ID, fetched.Spec.ID)
	}
}

// TestGetHardwareNotFound verifies that GET /hardwares/{uid} for an unknown
// UID returns HTTP 404.
func TestGetHardwareNotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/hardwares/hardware-does-not-exist", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 for unknown hardware, got %d", resp.StatusCode)
	}
}

// TestUpdateHardware verifies that PUT /hardwares/{uid} returns HTTP 200 and
// that the updated fields are persisted.
func TestUpdateHardware(t *testing.T) {
	created := createHardwareAndRequire(t, newHardware("x3000c0s5b0n3", "Node"))
	defer func() { doRequest(t, http.MethodDelete, "/hardwares/"+created.Metadata.UID, nil).Body.Close() }()

	updateReq := updateHardwareRequest{
		Metadata: componentMetadata{Name: "x3000c0s5b0n3"},
		Spec: hardwareSpec{
			ID:     "x3000c0s5b0n3",
			Type:   "Node",
			Status: "Empty",
		},
	}

	resp := doRequest(t, http.MethodPut, "/hardwares/"+created.Metadata.UID, updateReq)
	requireStatus(t, resp, http.StatusOK)

	var updated hardwareResponse
	decodeJSON(t, resp, &updated)

	if updated.Spec.Status != "Empty" {
		t.Errorf("expected Spec.Status=Empty after update, got %q", updated.Spec.Status)
	}

	// Confirm the change persisted via a subsequent GET
	getResp := doRequest(t, http.MethodGet, "/hardwares/"+created.Metadata.UID, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched hardwareResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.Status != "Empty" {
		t.Errorf("GET after PUT: expected Spec.Status=Empty, got %q", fetched.Spec.Status)
	}
}

// TestDeleteHardware verifies that DELETE /hardwares/{uid} returns HTTP 200
// and that a subsequent GET returns HTTP 404.
func TestDeleteHardware(t *testing.T) {
	created := createHardwareAndRequire(t, newHardware("x3000c0s5b0n4", "Node"))

	// Delete
	delResp := doRequest(t, http.MethodDelete, "/hardwares/"+created.Metadata.UID, nil)
	requireStatus(t, delResp, http.StatusOK)

	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.UID != created.Metadata.UID {
		t.Errorf("delete response UID mismatch: want %q, got %q", created.Metadata.UID, delBody.UID)
	}

	// Verify it is gone
	getResp := doRequest(t, http.MethodGet, "/hardwares/"+created.Metadata.UID, nil)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", getResp.StatusCode)
	}
}

// TestHardwareLifecycle exercises the full POST → GET → PUT → DELETE cycle
// in one test so the end-to-end flow is clearly visible.
func TestHardwareLifecycle(t *testing.T) {
	xname := "x3000c0s5b0n5"

	// ── POST ──────────────────────────────────────────────────────────────────
	created := createHardwareAndRequire(t, newHardware(xname, "Node"))
	uid := created.Metadata.UID
	t.Logf("Created hardware UID: %s", uid)

	// ── GET by UID ────────────────────────────────────────────────────────────
	getResp := doRequest(t, http.MethodGet, "/hardwares/"+uid, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched hardwareResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.ID != xname {
		t.Errorf("GET: expected Spec.ID=%q, got %q", xname, fetched.Spec.ID)
	}

	// ── GET list – hardware must appear ──────────────────────────────────────
	listResp := doRequest(t, http.MethodGet, "/hardwares", nil)
	requireStatus(t, listResp, http.StatusOK)
	var list []json.RawMessage
	decodeJSON(t, listResp, &list)
	found := false
	for _, raw := range list {
		var h hardwareResponse
		if err := json.Unmarshal(raw, &h); err == nil && h.Metadata.UID == uid {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("hardware %s not found in GET /hardwares list", uid)
	}

	// ── PUT ───────────────────────────────────────────────────────────────────
	putResp := doRequest(t, http.MethodPut, "/hardwares/"+uid, updateHardwareRequest{
		Metadata: componentMetadata{Name: xname},
		Spec:     hardwareSpec{ID: xname, Type: "Node", Status: "Populated"},
	})
	requireStatus(t, putResp, http.StatusOK)
	var updated hardwareResponse
	decodeJSON(t, putResp, &updated)
	if updated.Spec.Status != "Populated" {
		t.Errorf("PUT: expected Spec.Status=Populated, got %q", updated.Spec.Status)
	}

	// ── DELETE ────────────────────────────────────────────────────────────────
	delResp := doRequest(t, http.MethodDelete, "/hardwares/"+uid, nil)
	requireStatus(t, delResp, http.StatusOK)

	// Confirm deletion
	gone := doRequest(t, http.MethodGet, "/hardwares/"+uid, nil)
	defer gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", gone.StatusCode)
	}
}

// TestCreateHardwareDuplicateID verifies that POST /hardwares rejects a second
// resource with the same Spec.ID, enforcing resource_id uniqueness.
func TestCreateHardwareDuplicateID(t *testing.T) {
	xname := "x3001c0s5b0n0"
	first := createHardwareAndRequire(t, newHardware(xname, "Node"))
	defer func() { doRequest(t, http.MethodDelete, "/hardwares/"+first.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodPost, "/hardwares", newHardware(xname, "Node"))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate hardware ID %q, got HTTP %d", xname, resp.StatusCode)
	}
}
