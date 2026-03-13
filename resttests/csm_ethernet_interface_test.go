/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /hsm/v2/Inventory/EthernetInterfaces routes registered in csm_routes.go.
 * Routes under test:
 *   GET    /hsm/v2/Inventory/EthernetInterfaces
 *   POST   /hsm/v2/Inventory/EthernetInterfaces
 *   GET    /hsm/v2/Inventory/EthernetInterfaces/{id}
 *   PUT    /hsm/v2/Inventory/EthernetInterfaces/{id}
 *   DELETE /hsm/v2/Inventory/EthernetInterfaces/{id}
 *
 * Notes on request/response shapes (from csm_ethernet_interfaces.go):
 *   POST  body   : EthernetInterfaceSpec (single object)
 *   POST  returns: HTTP 201, echoes EthernetInterfaceSpec
 *   GET / returns: []*EthernetInterfaceSpec  (plain JSON array)
 *   GET /{id}    : EthernetInterfaceSpec (spec only)
 *   PUT  body   : EthernetInterfaceSpec
 *   PUT  returns: EthernetInterfaceSpec (updated)
 *   DELETE /{id}: DeleteResponse { message, uid }
 *   ID key      : EthernetInterfaceSpec.ID
 */

package resttests

import (
	"fmt"
	"net/http"
	"testing"
)

const csmEIBase = "/hsm/v2/Inventory/EthernetInterfaces"

// ─── CSM request / response shapes ───────────────────────────────────────────

type csmIPAddress struct {
	IPAddress string `json:"IPAddress"`
	Network   string `json:"Network,omitempty"`
}

type csmEthernetInterfaceSpec struct {
	ID          string         `json:"ID"`
	Description string         `json:"Description,omitempty"`
	MACAddress  string         `json:"MACAddress"`
	LastUpdate  string         `json:"LastUpdate,omitempty"`
	ComponentID string         `json:"ComponentID,omitempty"`
	Type        string         `json:"Type,omitempty"`
	IPAddresses []csmIPAddress `json:"IPAddresses,omitempty"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// csmEICreate POSTs a single EthernetInterface via the CSM endpoint and asserts 201.
func csmEICreate(t *testing.T, spec csmEthernetInterfaceSpec) {
	t.Helper()
	resp := doRequest(t, http.MethodPost, csmEIBase, spec)
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// csmEIGetOne fetches a single ethernet interface by ID.
func csmEIGetOne(t *testing.T, id string) (*csmEthernetInterfaceSpec, int) {
	t.Helper()
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("%s/%s", csmEIBase, id), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var spec csmEthernetInterfaceSpec
	decodeJSON(t, resp, &spec)
	return &spec, resp.StatusCode
}

// csmEIDelete deletes an ethernet interface by ID.
func csmEIDelete(t *testing.T, id string) {
	t.Helper()
	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmEIBase, id), nil)
	resp.Body.Close()
}

func newCsmEthernetInterface(id, mac, componentID string) csmEthernetInterfaceSpec {
	return csmEthernetInterfaceSpec{
		ID:          id,
		MACAddress:  mac,
		ComponentID: componentID,
		Type:        "HostLAN",
		IPAddresses: []csmIPAddress{{IPAddress: "10.0.1.1"}},
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateEthernetInterfaceCsm verifies POST /hsm/v2/Inventory/EthernetInterfaces returns 201.
func TestCreateEthernetInterfaceCsm(t *testing.T) {
	id := "b0:00:00:00:00:01"
	csmEICreate(t, newCsmEthernetInterface(id, id, "x3000c0s0b0n0"))
	defer csmEIDelete(t, id)

	spec, status := csmEIGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GET after POST, got %d", status)
	}
	if spec.MACAddress != id {
		t.Errorf("expected MACAddress=%q, got %q", id, spec.MACAddress)
	}
}

// TestGetEthernetInterfacesCsm verifies GET /hsm/v2/Inventory/EthernetInterfaces
// returns HTTP 200 and a plain array containing the created interface.
func TestGetEthernetInterfacesCsm(t *testing.T) {
	id := "b0:00:00:00:00:02"
	csmEICreate(t, newCsmEthernetInterface(id, id, "x3000c0s0b0n0"))
	defer csmEIDelete(t, id)

	resp := doRequest(t, http.MethodGet, csmEIBase, nil)
	requireStatus(t, resp, http.StatusOK)

	var list []csmEthernetInterfaceSpec
	decodeJSON(t, resp, &list)

	found := false
	for _, ei := range list {
		if ei.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ethernet interface %s not found in GET %s list", id, csmEIBase)
	}
}

// TestGetEthernetInterfaceCsm verifies GET /hsm/v2/Inventory/EthernetInterfaces/{id}
// returns 200 and the correct spec.
func TestGetEthernetInterfaceCsm(t *testing.T) {
	id := "b0:00:00:00:00:03"
	csmEICreate(t, newCsmEthernetInterface(id, id, "x3000c0s0b0n1"))
	defer csmEIDelete(t, id)

	spec, status := csmEIGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", status)
	}
	if spec.ID != id {
		t.Errorf("expected ID=%q, got %q", id, spec.ID)
	}
	if spec.ComponentID != "x3000c0s0b0n1" {
		t.Errorf("expected ComponentID=x3000c0s0b0n1, got %q", spec.ComponentID)
	}
}

// TestUpdateEthernetInterfaceCsm verifies PUT /hsm/v2/Inventory/EthernetInterfaces/{id}
// updates the spec and returns 200.
func TestUpdateEthernetInterfaceCsm(t *testing.T) {
	id := "b0:00:00:00:00:04"
	csmEICreate(t, newCsmEthernetInterface(id, id, "x3000c0s0b0n0"))
	defer csmEIDelete(t, id)

	updateSpec := csmEthernetInterfaceSpec{
		ID:          id,
		MACAddress:  id,
		ComponentID: "x3000c0s0b0n0",
		Type:        "HostLAN",
		IPAddresses: []csmIPAddress{{IPAddress: "10.0.99.99"}},
	}
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmEIBase, id), updateSpec)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	spec, status := csmEIGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 after PUT, got %d", status)
	}
	if len(spec.IPAddresses) == 0 || spec.IPAddresses[0].IPAddress != "10.0.99.99" {
		t.Errorf("expected IPAddresses[0]=10.0.99.99 after PUT")
	}
}

// TestDeleteEthernetInterfaceCsm verifies DELETE /hsm/v2/Inventory/EthernetInterfaces/{id}
// returns 200 and that a subsequent GET does not return 200.
func TestDeleteEthernetInterfaceCsm(t *testing.T) {
	id := "b0:00:00:00:00:05"
	csmEICreate(t, newCsmEthernetInterface(id, id, "x3000c0s0b0n0"))

	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmEIBase, id), nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.Message == "" {
		t.Error("expected non-empty message in delete response")
	}

	_, status := csmEIGetOne(t, id)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE %s, still got 200", id)
	}
}

// TestCsmEthernetInterfaceLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestCsmEthernetInterfaceLifecycle(t *testing.T) {
	id := "b0:00:00:00:00:06"
	componentID := "x3000c0s0b0n0"

	// POST
	csmEICreate(t, newCsmEthernetInterface(id, id, componentID))

	// GET
	spec, status := csmEIGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("POST→GET: expected HTTP 200, got %d", status)
	}
	if spec.ID != id {
		t.Errorf("POST→GET: expected ID=%q, got %q", id, spec.ID)
	}

	// GET all – must appear
	listResp := doRequest(t, http.MethodGet, csmEIBase, nil)
	requireStatus(t, listResp, http.StatusOK)
	var list []csmEthernetInterfaceSpec
	decodeJSON(t, listResp, &list)
	found := false
	for _, ei := range list {
		if ei.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ethernet interface %s not found in GET list after creation", id)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmEIBase, id), csmEthernetInterfaceSpec{
		ID: id, MACAddress: id, ComponentID: "x3000c0s0b0n2", Type: "HostLAN",
		IPAddresses: []csmIPAddress{{IPAddress: "10.5.6.7"}},
	})
	requireStatus(t, putResp, http.StatusOK)
	putResp.Body.Close()

	spec, status = csmEIGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("GET after PUT: expected HTTP 200, got %d", status)
	}
	if spec.ComponentID != "x3000c0s0b0n2" {
		t.Errorf("PUT: expected ComponentID=x3000c0s0b0n2, got %q", spec.ComponentID)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmEIBase, id), nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	_, status = csmEIGetOne(t, id)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE, still got 200")
	}
}

// TestCreateEthernetInterfaceCsmDuplicateID verifies that POST .../EthernetInterfaces
// rejects an ethernet interface whose ID already exists, enforcing resource_id uniqueness.
func TestCreateEthernetInterfaceCsmDuplicateID(t *testing.T) {
	id := "c0:00:00:00:00:02"
	csmEICreate(t, newCsmEthernetInterface(id, id, "x3000c0s0b0n0"))
	defer csmEIDelete(t, id)

	resp := doRequest(t, http.MethodPost, csmEIBase, newCsmEthernetInterface(id, id, "x3000c0s0b0n0"))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate ethernet interface ID %q, got HTTP %d", id, resp.StatusCode)
	}
}
