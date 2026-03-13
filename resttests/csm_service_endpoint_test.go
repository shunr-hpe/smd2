/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /hsm/v2/Inventory/ServiceEndpoints routes registered in csm_routes.go.
 * Routes under test:
 *   GET    /hsm/v2/Inventory/ServiceEndpoints
 *   POST   /hsm/v2/Inventory/ServiceEndpoints
 *   GET    /hsm/v2/Inventory/ServiceEndpoints/{id}
 *   PUT    /hsm/v2/Inventory/ServiceEndpoints/{id}
 *   DELETE /hsm/v2/Inventory/ServiceEndpoints/{id}
 *
 * Notes on request/response shapes (from csm_service_endpoints.go):
 *   POST  body   : ServiceEndpointArray { "ServiceEndpoints": [ <ServiceEndpointSpec>, ... ] }
 *   POST  returns: HTTP 201, echoes the ServiceEndpointArray
 *   GET / returns: ServiceEndpointArray { "ServiceEndpoints": [ <ServiceEndpointSpec>, ... ] }
 *   GET /{id}    : ServiceEndpointSpec (spec only)
 *   PUT  body   : ServiceEndpointSpec
 *   PUT  returns: ServiceEndpointSpec (updated)
 *   DELETE /{id}: DeleteResponse { message, uid }
 *   ID key      : ServiceEndpointSpec.RedfishEndpointID
 */

package resttests

import (
	"fmt"
	"net/http"
	"testing"
)

const csmSEBase = "/hsm/v2/Inventory/ServiceEndpoints"

// ─── CSM request / response shapes ───────────────────────────────────────────

type csmServiceEndpointSpec struct {
	RedfishEndpointID   string `json:"RedfishEndpointID"`
	RedfishType         string `json:"RedfishType,omitempty"`
	RedfishSubtype      string `json:"RedfishSubtype,omitempty"`
	UUID                string `json:"UUID,omitempty"`
	OdataID             string `json:"OdataID,omitempty"`
	RedfishEndpointFQDN string `json:"RedfishEndpointFQDN,omitempty"`
	RedfishURL          string `json:"RedfishURL,omitempty"`
}

type csmServiceEndpointArray struct {
	ServiceEndpoints []*csmServiceEndpointSpec `json:"ServiceEndpoints"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// csmSECreate POSTs a batch of service endpoints via the CSM endpoint and asserts 201.
func csmSECreate(t *testing.T, specs ...*csmServiceEndpointSpec) {
	t.Helper()
	body := csmServiceEndpointArray{ServiceEndpoints: specs}
	resp := doRequest(t, http.MethodPost, csmSEBase, body)
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// csmSEGetOne fetches a single service endpoint by RedfishEndpointID.
func csmSEGetOne(t *testing.T, rfID string) (*csmServiceEndpointSpec, int) {
	t.Helper()
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("%s/%s", csmSEBase, rfID), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var spec csmServiceEndpointSpec
	decodeJSON(t, resp, &spec)
	return &spec, resp.StatusCode
}

// csmSEDelete deletes a service endpoint by RedfishEndpointID.
func csmSEDelete(t *testing.T, rfID string) {
	t.Helper()
	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmSEBase, rfID), nil)
	resp.Body.Close()
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateServiceEndpointCsm verifies POST /hsm/v2/Inventory/ServiceEndpoints returns 201.
func TestCreateServiceEndpointCsm(t *testing.T) {
	rfID := "x3000c0r1b0"
	csmSECreate(t, &csmServiceEndpointSpec{
		RedfishEndpointID: rfID,
		RedfishType:       "Chassis",
		OdataID:           "/redfish/v1/Chassis/" + rfID,
	})
	defer csmSEDelete(t, rfID)

	spec, status := csmSEGetOne(t, rfID)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GET after POST, got %d", status)
	}
	if spec.RedfishEndpointID != rfID {
		t.Errorf("expected RedfishEndpointID=%q, got %q", rfID, spec.RedfishEndpointID)
	}
}

// TestCreateServiceEndpointCsmBulk verifies multiple service endpoints can be created in a single POST.
func TestCreateServiceEndpointCsmBulk(t *testing.T) {
	rfIDs := []string{"x3000c0r1b1", "x3000c0r1b2", "x3000c0r1b3"}
	specs := make([]*csmServiceEndpointSpec, len(rfIDs))
	for i, id := range rfIDs {
		specs[i] = &csmServiceEndpointSpec{RedfishEndpointID: id, RedfishType: "Manager"}
	}
	csmSECreate(t, specs...)
	defer func() {
		for _, id := range rfIDs {
			csmSEDelete(t, id)
		}
	}()

	for _, id := range rfIDs {
		spec, status := csmSEGetOne(t, id)
		if status != http.StatusOK {
			t.Errorf("expected HTTP 200 for %s, got %d", id, status)
			continue
		}
		if spec.RedfishEndpointID != id {
			t.Errorf("expected RedfishEndpointID=%q, got %q", id, spec.RedfishEndpointID)
		}
	}
}

// TestGetServiceEndpointsCsm verifies GET /hsm/v2/Inventory/ServiceEndpoints returns ServiceEndpointArray.
func TestGetServiceEndpointsCsm(t *testing.T) {
	rfID := "x3000c0r1b4"
	csmSECreate(t, &csmServiceEndpointSpec{RedfishEndpointID: rfID, RedfishType: "Chassis"})
	defer csmSEDelete(t, rfID)

	resp := doRequest(t, http.MethodGet, csmSEBase, nil)
	requireStatus(t, resp, http.StatusOK)

	var list csmServiceEndpointArray
	decodeJSON(t, resp, &list)

	found := false
	for _, s := range list.ServiceEndpoints {
		if s != nil && s.RedfishEndpointID == rfID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("service endpoint %s not found in GET %s list", rfID, csmSEBase)
	}
}

// TestGetServiceEndpointCsm verifies GET /hsm/v2/Inventory/ServiceEndpoints/{id} returns the spec.
func TestGetServiceEndpointCsm(t *testing.T) {
	rfID := "x3000c0r1b5"
	csmSECreate(t, &csmServiceEndpointSpec{RedfishEndpointID: rfID, RedfishType: "Manager"})
	defer csmSEDelete(t, rfID)

	spec, status := csmSEGetOne(t, rfID)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", status)
	}
	if spec.RedfishEndpointID != rfID {
		t.Errorf("expected RedfishEndpointID=%q, got %q", rfID, spec.RedfishEndpointID)
	}
}

// TestUpdateServiceEndpointCsm verifies PUT /hsm/v2/Inventory/ServiceEndpoints/{id}
// updates the resource and returns the updated spec.
func TestUpdateServiceEndpointCsm(t *testing.T) {
	rfID := "x3000c0r1b6"
	csmSECreate(t, &csmServiceEndpointSpec{
		RedfishEndpointID: rfID,
		RedfishType:       "Chassis",
		OdataID:           "/redfish/v1/Chassis/" + rfID,
	})
	defer csmSEDelete(t, rfID)

	updateSpec := csmServiceEndpointSpec{
		RedfishEndpointID:   rfID,
		RedfishType:         "Manager",
		RedfishEndpointFQDN: "bmc.example.com",
		OdataID:             "/redfish/v1/Managers/BMC",
	}
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmSEBase, rfID), updateSpec)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Verify via GET that the update persisted
	spec, status := csmSEGetOne(t, rfID)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 after PUT, got %d", status)
	}
	if spec.RedfishType != "Manager" {
		t.Errorf("expected RedfishType=Manager after PUT, got %q", spec.RedfishType)
	}
	if spec.RedfishEndpointFQDN != "bmc.example.com" {
		t.Errorf("expected RedfishEndpointFQDN=bmc.example.com after PUT, got %q", spec.RedfishEndpointFQDN)
	}
}

// TestDeleteServiceEndpointCsm verifies DELETE /hsm/v2/Inventory/ServiceEndpoints/{id}
// returns 200 and that a subsequent GET does not return 200.
func TestDeleteServiceEndpointCsm(t *testing.T) {
	rfID := "x3000c0r1b7"
	csmSECreate(t, &csmServiceEndpointSpec{RedfishEndpointID: rfID, RedfishType: "Chassis"})

	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmSEBase, rfID), nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.Message == "" {
		t.Error("expected non-empty message in delete response")
	}

	// Confirm gone
	_, status := csmSEGetOne(t, rfID)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE %s, still got 200", rfID)
	}
}

// TestCsmServiceEndpointLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestCsmServiceEndpointLifecycle(t *testing.T) {
	rfID := "x3000c0r1b8"

	// POST
	csmSECreate(t, &csmServiceEndpointSpec{
		RedfishEndpointID: rfID,
		RedfishType:       "Chassis",
		OdataID:           "/redfish/v1/Chassis/" + rfID,
	})

	// GET
	spec, status := csmSEGetOne(t, rfID)
	if status != http.StatusOK {
		t.Fatalf("POST→GET: expected HTTP 200, got %d", status)
	}
	if spec.RedfishEndpointID != rfID {
		t.Errorf("POST→GET: expected RedfishEndpointID=%q, got %q", rfID, spec.RedfishEndpointID)
	}

	// GET all – must appear
	listResp := doRequest(t, http.MethodGet, csmSEBase, nil)
	requireStatus(t, listResp, http.StatusOK)
	var list csmServiceEndpointArray
	decodeJSON(t, listResp, &list)
	found := false
	for _, s := range list.ServiceEndpoints {
		if s != nil && s.RedfishEndpointID == rfID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("service endpoint %s not found in GET list after creation", rfID)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmSEBase, rfID), csmServiceEndpointSpec{
		RedfishEndpointID:   rfID,
		RedfishType:         "Manager",
		RedfishEndpointFQDN: "updated.example.com",
		OdataID:             "/redfish/v1/Managers/BMC",
	})
	requireStatus(t, putResp, http.StatusOK)
	putResp.Body.Close()

	spec, status = csmSEGetOne(t, rfID)
	if status != http.StatusOK {
		t.Fatalf("GET after PUT: expected HTTP 200, got %d", status)
	}
	if spec.RedfishType != "Manager" {
		t.Errorf("PUT: expected RedfishType=Manager, got %q", spec.RedfishType)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmSEBase, rfID), nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	_, status = csmSEGetOne(t, rfID)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE, still got 200")
	}
}

// TestCreateServiceEndpointCsmDuplicateID verifies that POST .../ServiceEndpoints
// rejects a service endpoint whose RedfishURL already exists, enforcing resource_id
// uniqueness. resource_id for ServiceEndpoints is Spec.RedfishURL, so a non-empty
// RedfishURL is required for the uniqueness constraint to apply.
func TestCreateServiceEndpointCsmDuplicateID(t *testing.T) {
	const redfishURL = "x3001c0r0b0-mgr"
	spec := &csmServiceEndpointSpec{
		RedfishEndpointID: "x3001c0r0b0",
		RedfishType:       "Manager",
		RedfishURL:        redfishURL,
	}
	csmSECreate(t, spec)
	defer csmSEDelete(t, redfishURL) // resource_id equals RedfishURL

	resp := doRequest(t, http.MethodPost, csmSEBase, csmServiceEndpointArray{
		ServiceEndpoints: []*csmServiceEndpointSpec{spec},
	})
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate service endpoint URL %q, got HTTP %d", redfishURL, resp.StatusCode)
	}
}
