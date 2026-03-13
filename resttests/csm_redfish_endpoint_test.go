/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /hsm/v2/Inventory/RedfishEndpoints routes registered in csm_routes.go.
 * Routes under test:
 *   GET    /hsm/v2/Inventory/RedfishEndpoints
 *   POST   /hsm/v2/Inventory/RedfishEndpoints
 *   GET    /hsm/v2/Inventory/RedfishEndpoints/{id}
 *   PUT    /hsm/v2/Inventory/RedfishEndpoints/{id}
 *   DELETE /hsm/v2/Inventory/RedfishEndpoints/{id}
 *
 * Notes on request/response shapes (from csm_redfish_endpoints.go):
 *   POST  body   : RedfishEndpointSpec (single object, not an array)
 *   POST  returns: HTTP 201, returns RedfishEndpointSpec
 *   GET / returns: RedfishEndpointArray { "RedfishEndpoints": [ <RedfishEndpointSpec>, ... ] }
 *   GET /{id}    : RedfishEndpointSpec (spec only)
 *   PUT  body   : RedfishEndpointSpec
 *   PUT  returns: RedfishEndpointSpec (updated)
 *   DELETE /{id}: DeleteResponse { message, uid }
 *   ID key      : RedfishEndpointSpec.ID
 */

package resttests

import (
	"fmt"
	"net/http"
	"testing"
)

const csmREBase = "/hsm/v2/Inventory/RedfishEndpoints"

// ─── CSM request / response shapes ───────────────────────────────────────────

type csmRedfishEndpointSpec struct {
	ID                 string            `json:"ID"`
	Type               string            `json:"Type,omitempty"`
	Hostname           string            `json:"Hostname,omitempty"`
	Domain             string            `json:"Domain,omitempty"`
	FQDN               string            `json:"FQDN,omitempty"`
	Enabled            bool              `json:"Enabled,omitempty"`
	User               string            `json:"User,omitempty"`
	Password           string            `json:"Password,omitempty"`
	RediscoverOnUpdate bool              `json:"RediscoverOnUpdate,omitempty"`
	DiscoveryInfo      csmDiscoveryInfo  `json:"DiscoveryInfo,omitempty"`
}

type csmDiscoveryInfo struct {
	LastDiscoveryAttempt string `json:"LastDiscoveryAttempt,omitempty"`
	LastDiscoveryStatus  string `json:"LastDiscoveryStatus"`
	RedfishVersion       string `json:"RedfishVersion,omitempty"`
}

type csmRedfishEndpointArray struct {
	RedfishEndpoints []*csmRedfishEndpointSpec `json:"RedfishEndpoints"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// csmRECreate POSTs a single RedfishEndpoint via the CSM endpoint and asserts 201.
func csmRECreate(t *testing.T, spec csmRedfishEndpointSpec) {
	t.Helper()
	resp := doRequest(t, http.MethodPost, csmREBase, spec)
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// csmREGetOne fetches a single redfish endpoint by ID.
func csmREGetOne(t *testing.T, id string) (*csmRedfishEndpointSpec, int) {
	t.Helper()
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("%s/%s", csmREBase, id), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var spec csmRedfishEndpointSpec
	decodeJSON(t, resp, &spec)
	return &spec, resp.StatusCode
}

// csmREDelete deletes a redfish endpoint by ID.
func csmREDelete(t *testing.T, id string) {
	t.Helper()
	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmREBase, id), nil)
	resp.Body.Close()
}

func newCsmRedfishEndpoint(id, hostname string) csmRedfishEndpointSpec {
	return csmRedfishEndpointSpec{
		ID:       id,
		Type:     "NodeBMC",
		Hostname: hostname,
		FQDN:     hostname,
		Enabled:  true,
		User:     "root",
		Password: "password",
		DiscoveryInfo: csmDiscoveryInfo{
			LastDiscoveryStatus: "NotYetQueried",
		},
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateRedfishEndpointCsm verifies POST /hsm/v2/Inventory/RedfishEndpoints returns 201.
func TestCreateRedfishEndpointCsm(t *testing.T) {
	id := "x3000c0s7b0"
	csmRECreate(t, newCsmRedfishEndpoint(id, "bmc7.example.com"))
	defer csmREDelete(t, id)

	spec, status := csmREGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GET after POST, got %d", status)
	}
	if spec.ID != id {
		t.Errorf("expected ID=%q, got %q", id, spec.ID)
	}
}

// TestGetRedfishEndpointsCsm verifies GET /hsm/v2/Inventory/RedfishEndpoints
// returns HTTP 200 and a RedfishEndpointArray containing the created endpoint.
func TestGetRedfishEndpointsCsm(t *testing.T) {
	id := "x3000c0s8b0"
	csmRECreate(t, newCsmRedfishEndpoint(id, "bmc8.example.com"))
	defer csmREDelete(t, id)

	resp := doRequest(t, http.MethodGet, csmREBase, nil)
	requireStatus(t, resp, http.StatusOK)

	var list csmRedfishEndpointArray
	decodeJSON(t, resp, &list)

	found := false
	for _, re := range list.RedfishEndpoints {
		if re != nil && re.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("redfish endpoint %s not found in GET %s list", id, csmREBase)
	}
}

// TestGetRedfishEndpointCsm verifies GET /hsm/v2/Inventory/RedfishEndpoints/{id}
// returns 200 and the correct spec.
func TestGetRedfishEndpointCsm(t *testing.T) {
	id := "x3000c0s9b0"
	csmRECreate(t, newCsmRedfishEndpoint(id, "bmc9.example.com"))
	defer csmREDelete(t, id)

	spec, status := csmREGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", status)
	}
	if spec.ID != id {
		t.Errorf("expected ID=%q, got %q", id, spec.ID)
	}
	if spec.Hostname != "bmc9.example.com" {
		t.Errorf("expected Hostname=bmc9.example.com, got %q", spec.Hostname)
	}
}

// TestUpdateRedfishEndpointCsm verifies PUT /hsm/v2/Inventory/RedfishEndpoints/{id}
// updates the spec and returns 200.
func TestUpdateRedfishEndpointCsm(t *testing.T) {
	id := "x3000c0s10b0"
	csmRECreate(t, newCsmRedfishEndpoint(id, "bmc10.example.com"))
	defer csmREDelete(t, id)

	updateSpec := csmRedfishEndpointSpec{
		ID:       id,
		Type:     "NodeBMC",
		Hostname: "bmc10-updated.example.com",
		FQDN:     "bmc10-updated.example.com",
		Enabled:  true,
		User:     "admin",
		Password: "newpass",
		DiscoveryInfo: csmDiscoveryInfo{
			LastDiscoveryStatus: "DiscoverOK",
		},
	}
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmREBase, id), updateSpec)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	spec, status := csmREGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 after PUT, got %d", status)
	}
	if spec.Hostname != "bmc10-updated.example.com" {
		t.Errorf("expected Hostname=bmc10-updated.example.com after PUT, got %q", spec.Hostname)
	}
	if spec.User != "admin" {
		t.Errorf("expected User=admin after PUT, got %q", spec.User)
	}
}

// TestDeleteRedfishEndpointCsm verifies DELETE /hsm/v2/Inventory/RedfishEndpoints/{id}
// returns 200 and that a subsequent GET does not return 200.
func TestDeleteRedfishEndpointCsm(t *testing.T) {
	id := "x3000c0s11b0"
	csmRECreate(t, newCsmRedfishEndpoint(id, "bmc11.example.com"))

	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmREBase, id), nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.Message == "" {
		t.Error("expected non-empty message in delete response")
	}

	_, status := csmREGetOne(t, id)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE %s, still got 200", id)
	}
}

// TestCsmRedfishEndpointLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestCsmRedfishEndpointLifecycle(t *testing.T) {
	id := "x3000c0s12b0"

	// POST
	csmRECreate(t, newCsmRedfishEndpoint(id, "bmc12.example.com"))

	// GET
	spec, status := csmREGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("POST→GET: expected HTTP 200, got %d", status)
	}
	if spec.ID != id {
		t.Errorf("POST→GET: expected ID=%q, got %q", id, spec.ID)
	}

	// GET all – must appear
	listResp := doRequest(t, http.MethodGet, csmREBase, nil)
	requireStatus(t, listResp, http.StatusOK)
	var list csmRedfishEndpointArray
	decodeJSON(t, listResp, &list)
	found := false
	for _, re := range list.RedfishEndpoints {
		if re != nil && re.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("redfish endpoint %s not found in GET list after creation", id)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmREBase, id), csmRedfishEndpointSpec{
		ID: id, Type: "NodeBMC",
		Hostname: "bmc12-new.example.com", FQDN: "bmc12-new.example.com",
		Enabled: true, User: "root", Password: "pass",
		DiscoveryInfo: csmDiscoveryInfo{LastDiscoveryStatus: "DiscoverOK"},
	})
	requireStatus(t, putResp, http.StatusOK)
	putResp.Body.Close()

	spec, status = csmREGetOne(t, id)
	if status != http.StatusOK {
		t.Fatalf("GET after PUT: expected HTTP 200, got %d", status)
	}
	if spec.Hostname != "bmc12-new.example.com" {
		t.Errorf("PUT: expected Hostname=bmc12-new.example.com, got %q", spec.Hostname)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmREBase, id), nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	_, status = csmREGetOne(t, id)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE, still got 200")
	}
}

// TestCreateRedfishEndpointCsmDuplicateID verifies that POST .../RedfishEndpoints
// rejects a redfish endpoint whose ID already exists, enforcing resource_id uniqueness.
func TestCreateRedfishEndpointCsmDuplicateID(t *testing.T) {
	id := "x3000c0s13b0"
	csmRECreate(t, newCsmRedfishEndpoint(id, "bmc-dup.example.com"))
	defer csmREDelete(t, id)

	resp := doRequest(t, http.MethodPost, csmREBase, newCsmRedfishEndpoint(id, "bmc-dup.example.com"))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate redfish endpoint ID %q, got HTTP %d", id, resp.StatusCode)
	}
}
