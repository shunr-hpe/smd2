/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /hsm/v2/Inventory/ComponentEndpoints routes registered in csm_routes.go.
 * Routes under test:
 *   GET    /hsm/v2/Inventory/ComponentEndpoints
 *   POST   /hsm/v2/Inventory/ComponentEndpoints
 *   GET    /hsm/v2/Inventory/ComponentEndpoints/{id}
 *   PUT    /hsm/v2/Inventory/ComponentEndpoints/{id}
 *   DELETE /hsm/v2/Inventory/ComponentEndpoints/{id}
 *
 * Notes on request/response shapes (from csm_component_endpoints.go):
 *   POST  body   : ComponentEndpointArray { "ComponentEndpoints": [ <ComponentEndpointSpec>, ... ] }
 *   POST  returns: HTTP 201, echoes the ComponentEndpointArray
 *   GET / returns: ComponentEndpointArray { "ComponentEndpoints": [ <ComponentEndpointSpec>, ... ] }
 *   GET /{id}    : ComponentEndpointSpec (spec only)
 *   PUT  body   : ComponentEndpointSpec
 *   PUT  returns: ComponentEndpointSpec (updated)
 *   DELETE /{id}: DeleteResponse { message, uid }
 *   ID key      : ComponentEndpointSpec.ID
 */

package resttests

import (
	"fmt"
	"net/http"
	"testing"
)

const csmCEBase = "/hsm/v2/Inventory/ComponentEndpoints"

// ─── CSM request / response shapes ───────────────────────────────────────────

type csmComponentEndpointSpec struct {
	ID                    string `json:"ID"`
	Type                  string `json:"Type,omitempty"`
	RedfishType           string `json:"RedfishType,omitempty"`
	RedfishSubtype        string `json:"RedfishSubtype,omitempty"`
	OdataID               string `json:"OdataID,omitempty"`
	RedfishEndpointID     string `json:"RedfishEndpointID,omitempty"`
	RedfishEndpointFQDN   string `json:"RedfishEndpointFQDN,omitempty"`
	RedfishURL            string `json:"RedfishURL,omitempty"`
	ComponentEndpointType string `json:"ComponentEndpointType,omitempty"`
	Enabled               bool   `json:"Enabled,omitempty"`
}

type csmComponentEndpointArray struct {
	ComponentEndpoints []*csmComponentEndpointSpec `json:"ComponentEndpoints"`
}

// newCsmComponentEndpoint builds a spec suitable for bulk-POST.
func newCsmComponentEndpoint(id, rfEndpointID string) *csmComponentEndpointSpec {
	return &csmComponentEndpointSpec{
		ID:                    id,
		Type:                  "Node",
		RedfishType:           "ComputerSystem",
		RedfishSubtype:        "Physical",
		OdataID:               "/redfish/v1/Systems/1",
		RedfishEndpointID:     rfEndpointID,
		RedfishEndpointFQDN:   rfEndpointID + ".example.com",
		RedfishURL:            rfEndpointID + ".example.com/redfish/v1/Systems/1",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		Enabled:               true,
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// csmCECreate POSTs a batch of component endpoints via the CSM endpoint and asserts 201.
func csmCECreate(t *testing.T, specs ...*csmComponentEndpointSpec) {
	t.Helper()
	body := csmComponentEndpointArray{ComponentEndpoints: specs}
	resp := doRequest(t, http.MethodPost, csmCEBase, body)
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// csmCEGetOne fetches a single component endpoint by ID.
func csmCEGetOne(t *testing.T, id string) (*csmComponentEndpointSpec, int) {
	t.Helper()
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("%s/%s", csmCEBase, id), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var spec csmComponentEndpointSpec
	decodeJSON(t, resp, &spec)
	return &spec, resp.StatusCode
}

// csmCEDelete deletes a component endpoint by ID.
func csmCEDelete(t *testing.T, id string) {
	t.Helper()
	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmCEBase, id), nil)
	resp.Body.Close()
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateComponentEndpointCsm verifies POST /hsm/v2/Inventory/ComponentEndpoints returns 201.
func TestCreateComponentEndpointCsm(t *testing.T) {
	id := "x5000c0s0b0n0"
	csmCECreate(t, newCsmComponentEndpoint(id, "x5000c0s0b0"))
	defer csmCEDelete(t, id)

	spec, status := csmCEGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GET after POST, got %d", status)
	}
	if spec.ID != id {
		t.Errorf("expected ID=%q, got %q", id, spec.ID)
	}
}

// TestCreateComponentEndpointCsmBulk verifies multiple component endpoints can be created in a single POST.
func TestCreateComponentEndpointCsmBulk(t *testing.T) {
	ids := []string{"x5000c0s1b0n0", "x5000c0s1b0n1", "x5000c0s1b0n2"}
	specs := make([]*csmComponentEndpointSpec, len(ids))
	for i, id := range ids {
		specs[i] = newCsmComponentEndpoint(id, "x5000c0s1b0")
	}
	csmCECreate(t, specs...)
	defer func() {
		for _, id := range ids {
			csmCEDelete(t, id)
		}
	}()

	for _, id := range ids {
		spec, status := csmCEGetOne(t, id)
		if status != http.StatusOK {
			t.Errorf("expected HTTP 200 for %s, got %d", id, status)
			continue
		}
		if spec.ID != id {
			t.Errorf("expected ID=%q, got %q", id, spec.ID)
		}
	}
}

// TestGetComponentEndpointsCsm verifies GET /hsm/v2/Inventory/ComponentEndpoints
// returns HTTP 200 and a ComponentEndpointArray containing the created endpoint.
func TestGetComponentEndpointsCsm(t *testing.T) {
	id := "x5000c0s2b0n0"
	csmCECreate(t, newCsmComponentEndpoint(id, "x5000c0s2b0"))
	defer csmCEDelete(t, id)

	resp := doRequest(t, http.MethodGet, csmCEBase, nil)
	requireStatus(t, resp, http.StatusOK)

	var arr csmComponentEndpointArray
	decodeJSON(t, resp, &arr)

	found := false
	for _, ce := range arr.ComponentEndpoints {
		if ce != nil && ce.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("component endpoint %s not found in GET %s list", id, csmCEBase)
	}
}

// TestGetComponentEndpointCsm verifies GET /hsm/v2/Inventory/ComponentEndpoints/{id}
// returns 200 and the correct spec.
func TestGetComponentEndpointCsm(t *testing.T) {
	id := "x5000c0s3b0n0"
	csmCECreate(t, newCsmComponentEndpoint(id, "x5000c0s3b0"))
	defer csmCEDelete(t, id)

	spec, status := csmCEGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", status)
	}
	if spec.ID != id {
		t.Errorf("expected ID=%q, got %q", id, spec.ID)
	}
	if spec.RedfishEndpointID != "x5000c0s3b0" {
		t.Errorf("expected RedfishEndpointID=x5000c0s3b0, got %q", spec.RedfishEndpointID)
	}
}

// TestUpdateComponentEndpointCsm verifies PUT /hsm/v2/Inventory/ComponentEndpoints/{id}
// updates the spec and returns 200.
func TestUpdateComponentEndpointCsm(t *testing.T) {
	id := "x5000c0s4b0n0"
	csmCECreate(t, newCsmComponentEndpoint(id, "x5000c0s4b0"))
	defer csmCEDelete(t, id)

	updateSpec := csmComponentEndpointSpec{
		ID:                    id,
		Type:                  "Node",
		RedfishType:           "ComputerSystem",
		RedfishSubtype:        "Virtual",
		OdataID:               "/redfish/v1/Systems/1",
		RedfishEndpointID:     "x5000c0s4b0",
		RedfishEndpointFQDN:   "x5000c0s4b0.example.com",
		RedfishURL:            "x5000c0s4b0.example.com/redfish/v1/Systems/1",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		Enabled:               true,
	}
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmCEBase, id), updateSpec)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Verify via GET that the update persisted
	spec, status := csmCEGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 after PUT, got %d", status)
	}
	if spec.RedfishSubtype != "Virtual" {
		t.Errorf("expected RedfishSubtype=Virtual after PUT, got %q", spec.RedfishSubtype)
	}
}

// TestDeleteComponentEndpointCsm verifies DELETE /hsm/v2/Inventory/ComponentEndpoints/{id}
// returns 200 and that a subsequent GET does not return 200.
func TestDeleteComponentEndpointCsm(t *testing.T) {
	id := "x5000c0s5b0n0"
	csmCECreate(t, newCsmComponentEndpoint(id, "x5000c0s5b0"))

	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmCEBase, id), nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.Message == "" {
		t.Error("expected non-empty message in delete response")
	}

	// Confirm gone
	_, status := csmCEGetOne(t, id)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE %s, still got 200", id)
	}
}

// TestCsmComponentEndpointLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestCsmComponentEndpointLifecycle(t *testing.T) {
	id := "x5000c0s6b0n0"
	rfID := "x5000c0s6b0"

	// POST
	csmCECreate(t, newCsmComponentEndpoint(id, rfID))

	// GET
	spec, status := csmCEGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GET after POST, got %d", status)
	}
	if spec.ID != id {
		t.Errorf("expected ID=%q, got %q", id, spec.ID)
	}

	// PUT
	updateSpec := *spec
	updateSpec.RedfishSubtype = "Blade"
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmCEBase, id), updateSpec)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	got, status := csmCEGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 after PUT, got %d", status)
	}
	if got.RedfishSubtype != "Blade" {
		t.Errorf("expected RedfishSubtype=Blade after PUT, got %q", got.RedfishSubtype)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmCEBase, id), nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	// Confirm gone
	_, status = csmCEGetOne(t, id)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE %s, still got 200", id)
	}
}

// TestCreateComponentEndpointCsmDuplicateID verifies that POST .../ComponentEndpoints
// rejects a component endpoint whose ID already exists, enforcing resource_id uniqueness.
func TestCreateComponentEndpointCsmDuplicateID(t *testing.T) {
	id := "x5001c0s0b0n0"
	csmCECreate(t, newCsmComponentEndpoint(id, "x5001c0s0b0"))
	defer csmCEDelete(t, id)

	resp := doRequest(t, http.MethodPost, csmCEBase, csmComponentEndpointArray{
		ComponentEndpoints: []*csmComponentEndpointSpec{newCsmComponentEndpoint(id, "x5001c0s0b0")},
	})
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate component endpoint ID %q, got HTTP %d", id, resp.StatusCode)
	}
}
