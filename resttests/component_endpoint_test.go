/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /componentendpoints routes registered in routes_generated.go.
 * Routes under test:
 *   GET    /componentendpoints
 *   POST   /componentendpoints
 *   GET    /componentendpoints/{uid}
 *   PUT    /componentendpoints/{uid}
 *   DELETE /componentendpoints/{uid}
 */

package resttests

import (
	"net/http"
	"testing"
)

// ─── Request / response shapes ────────────────────────────────────────────────

type componentEndpointSpec struct {
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

type componentEndpointResponse struct {
	APIVersion string                `json:"apiVersion"`
	Kind       string                `json:"kind"`
	Metadata   componentMetadata     `json:"metadata"`
	Spec       componentEndpointSpec `json:"spec"`
}

type createComponentEndpointRequest struct {
	Metadata componentMetadata     `json:"metadata"`
	Spec     componentEndpointSpec `json:"spec"`
}

type updateComponentEndpointRequest struct {
	Metadata componentMetadata     `json:"metadata,omitempty"`
	Spec     componentEndpointSpec `json:"spec,omitempty"`
}

// newComponentEndpoint builds a valid create request.
func newComponentEndpoint(id, rfEndpointID string) createComponentEndpointRequest {
	return createComponentEndpointRequest{
		Metadata: componentMetadata{Name: id},
		Spec: componentEndpointSpec{
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
		},
	}
}

// createComponentEndpointAndRequire POSTs a component endpoint and validates 201.
func createComponentEndpointAndRequire(t *testing.T, req createComponentEndpointRequest) componentEndpointResponse {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/componentendpoints", req)
	requireStatus(t, resp, http.StatusCreated)
	var created componentEndpointResponse
	decodeJSON(t, resp, &created)
	if created.Metadata.UID == "" {
		t.Fatal("expected non-empty UID in created component endpoint response")
	}
	return created
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateComponentEndpoint verifies POST /componentendpoints returns 201 with a generated UID.
func TestCreateComponentEndpoint(t *testing.T) {
	req := newComponentEndpoint("x4000c0s0b0n0", "x4000c0s0b0")
	ce := createComponentEndpointAndRequire(t, req)

	if ce.Kind != "ComponentEndpoint" {
		t.Errorf("expected Kind=ComponentEndpoint, got %q", ce.Kind)
	}
	if ce.Spec.ID != req.Spec.ID {
		t.Errorf("expected Spec.ID=%q, got %q", req.Spec.ID, ce.Spec.ID)
	}

	doRequest(t, http.MethodDelete, "/componentendpoints/"+ce.Metadata.UID, nil).Body.Close()
}

// TestGetComponentEndpoints verifies GET /componentendpoints returns 200 and a non-empty array.
func TestGetComponentEndpoints(t *testing.T) {
	created := createComponentEndpointAndRequire(t, newComponentEndpoint("x4000c0s1b0n0", "x4000c0s1b0"))
	defer func() { doRequest(t, http.MethodDelete, "/componentendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/componentendpoints", nil)
	requireStatus(t, resp, http.StatusOK)

	var list []componentEndpointResponse
	decodeJSON(t, resp, &list)
	if len(list) == 0 {
		t.Error("expected at least one component endpoint in list, got zero")
	}
}

// TestGetComponentEndpoint verifies GET /componentendpoints/{uid} returns 200 and the correct resource.
func TestGetComponentEndpoint(t *testing.T) {
	created := createComponentEndpointAndRequire(t, newComponentEndpoint("x4000c0s2b0n0", "x4000c0s2b0"))
	defer func() { doRequest(t, http.MethodDelete, "/componentendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/componentendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, resp, http.StatusOK)

	var fetched componentEndpointResponse
	decodeJSON(t, resp, &fetched)
	if fetched.Metadata.UID != created.Metadata.UID {
		t.Errorf("expected UID=%q, got %q", created.Metadata.UID, fetched.Metadata.UID)
	}
	if fetched.Spec.ID != created.Spec.ID {
		t.Errorf("expected Spec.ID=%q, got %q", created.Spec.ID, fetched.Spec.ID)
	}
}

// TestGetComponentEndpointNotFound verifies GET /componentendpoints/{uid} for an unknown UID returns 404.
func TestGetComponentEndpointNotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/componentendpoints/componentendpoint-does-not-exist", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 for unknown component endpoint, got %d", resp.StatusCode)
	}
}

// TestUpdateComponentEndpoint verifies PUT /componentendpoints/{uid} returns 200 and persists changes.
func TestUpdateComponentEndpoint(t *testing.T) {
	created := createComponentEndpointAndRequire(t, newComponentEndpoint("x4000c0s3b0n0", "x4000c0s3b0"))
	defer func() { doRequest(t, http.MethodDelete, "/componentendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	updateReq := updateComponentEndpointRequest{
		Metadata: componentMetadata{Name: "x4000c0s3b0n0"},
		Spec: componentEndpointSpec{
			ID:                    "x4000c0s3b0n0",
			Type:                  "Node",
			RedfishType:           "ComputerSystem",
			RedfishSubtype:        "Virtual",
			OdataID:               "/redfish/v1/Systems/1",
			RedfishEndpointID:     "x4000c0s3b0",
			RedfishEndpointFQDN:   "x4000c0s3b0.example.com",
			RedfishURL:            "x4000c0s3b0.example.com/redfish/v1/Systems/1",
			ComponentEndpointType: "ComponentEndpointComputerSystem",
			Enabled:               true,
		},
	}
	resp := doRequest(t, http.MethodPut, "/componentendpoints/"+created.Metadata.UID, updateReq)
	requireStatus(t, resp, http.StatusOK)
	var updated componentEndpointResponse
	decodeJSON(t, resp, &updated)

	if updated.Spec.RedfishSubtype != "Virtual" {
		t.Errorf("expected Spec.RedfishSubtype=Virtual after PUT, got %q", updated.Spec.RedfishSubtype)
	}

	// Confirm via GET
	getResp := doRequest(t, http.MethodGet, "/componentendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched componentEndpointResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.RedfishSubtype != "Virtual" {
		t.Errorf("GET after PUT: expected Spec.RedfishSubtype=Virtual, got %q", fetched.Spec.RedfishSubtype)
	}
}

// TestDeleteComponentEndpoint verifies DELETE /componentendpoints/{uid} returns 200 and subsequent GET returns 404.
func TestDeleteComponentEndpoint(t *testing.T) {
	created := createComponentEndpointAndRequire(t, newComponentEndpoint("x4000c0s4b0n0", "x4000c0s4b0"))

	delResp := doRequest(t, http.MethodDelete, "/componentendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.UID != created.Metadata.UID {
		t.Errorf("delete response UID mismatch: want %q, got %q", created.Metadata.UID, delBody.UID)
	}

	getResp := doRequest(t, http.MethodGet, "/componentendpoints/"+created.Metadata.UID, nil)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", getResp.StatusCode)
	}
}

// TestComponentEndpointLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestComponentEndpointLifecycle(t *testing.T) {
	id := "x4000c0s5b0n0"
	rfID := "x4000c0s5b0"

	// POST
	created := createComponentEndpointAndRequire(t, newComponentEndpoint(id, rfID))
	uid := created.Metadata.UID
	t.Logf("Created component endpoint UID: %s", uid)

	// GET by UID
	getResp := doRequest(t, http.MethodGet, "/componentendpoints/"+uid, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched componentEndpointResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.ID != id {
		t.Errorf("GET: expected Spec.ID=%q, got %q", id, fetched.Spec.ID)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, "/componentendpoints/"+uid, updateComponentEndpointRequest{
		Metadata: componentMetadata{Name: id},
		Spec: componentEndpointSpec{
			ID:                    id,
			Type:                  "Node",
			RedfishType:           "ComputerSystem",
			RedfishSubtype:        "Blade",
			OdataID:               "/redfish/v1/Systems/1",
			RedfishEndpointID:     rfID,
			RedfishEndpointFQDN:   rfID + ".example.com",
			RedfishURL:            rfID + ".example.com/redfish/v1/Systems/1",
			ComponentEndpointType: "ComponentEndpointComputerSystem",
			Enabled:               true,
		},
	})
	requireStatus(t, putResp, http.StatusOK)
	var updated componentEndpointResponse
	decodeJSON(t, putResp, &updated)
	if updated.Spec.RedfishSubtype != "Blade" {
		t.Errorf("PUT: expected Spec.RedfishSubtype=Blade, got %q", updated.Spec.RedfishSubtype)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, "/componentendpoints/"+uid, nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	gone := doRequest(t, http.MethodGet, "/componentendpoints/"+uid, nil)
	defer gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", gone.StatusCode)
	}
}

// TestCreateComponentEndpointDuplicateID verifies that POST /componentendpoints rejects
// a second resource with the same Spec.ID, enforcing resource_id uniqueness.
func TestCreateComponentEndpointDuplicateID(t *testing.T) {
	id := "x4001c0s0b0n0"
	rfID := "x4001c0s0b0"
	first := createComponentEndpointAndRequire(t, newComponentEndpoint(id, rfID))
	defer func() { doRequest(t, http.MethodDelete, "/componentendpoints/"+first.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodPost, "/componentendpoints", newComponentEndpoint(id, rfID))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate component endpoint ID %q, got HTTP %d", id, resp.StatusCode)
	}
}
