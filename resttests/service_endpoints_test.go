/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /serviceendpoints routes registered in routes_generated.go.
 * Routes under test:
 *   GET    /serviceendpoints
 *   POST   /serviceendpoints
 *   GET    /serviceendpoints/{uid}
 *   PUT    /serviceendpoints/{uid}
 *   DELETE /serviceendpoints/{uid}
 */

package resttests

import (
	"net/http"
	"testing"
)

// ─── Request / response shapes ────────────────────────────────────────────────

type serviceEndpointSpec struct {
	RedfishEndpointID   string `json:"RedfishEndpointID"`
	RedfishType         string `json:"RedfishType,omitempty"`
	RedfishSubtype      string `json:"RedfishSubtype,omitempty"`
	UUID                string `json:"UUID,omitempty"`
	OdataID             string `json:"OdataID,omitempty"`
	RedfishEndpointFQDN string `json:"RedfishEndpointFQDN,omitempty"`
	RedfishURL          string `json:"RedfishURL,omitempty"`
}

type serviceEndpointResponse struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   componentMetadata   `json:"metadata"`
	Spec       serviceEndpointSpec `json:"spec"`
}

type createServiceEndpointRequest struct {
	Metadata componentMetadata   `json:"metadata"`
	Spec     serviceEndpointSpec `json:"spec"`
}

type updateServiceEndpointRequest struct {
	Metadata componentMetadata   `json:"metadata,omitempty"`
	Spec     serviceEndpointSpec `json:"spec,omitempty"`
}

// newServiceEndpoint builds a valid create request.
func newServiceEndpoint(rfEndpointID, rfType string) createServiceEndpointRequest {
	return createServiceEndpointRequest{
		Metadata: componentMetadata{Name: rfEndpointID},
		Spec: serviceEndpointSpec{
			RedfishEndpointID: rfEndpointID,
			RedfishType:       rfType,
			OdataID:           "/redfish/v1/Systems/" + rfEndpointID,
		},
	}
}

// createServiceEndpointAndRequire POSTs a service endpoint and validates 201.
func createServiceEndpointAndRequire(t *testing.T, req createServiceEndpointRequest) serviceEndpointResponse {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/serviceendpoints", req)
	requireStatus(t, resp, http.StatusCreated)
	var created serviceEndpointResponse
	decodeJSON(t, resp, &created)
	if created.Metadata.UID == "" {
		t.Fatal("expected non-empty UID in created service endpoint response")
	}
	return created
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateServiceEndpoint verifies POST /serviceendpoints returns 201 with a generated UID.
func TestCreateServiceEndpoint(t *testing.T) {
	req := newServiceEndpoint("x3000c0b0", "Chassis")
	se := createServiceEndpointAndRequire(t, req)

	if se.Kind != "ServiceEndpoint" {
		t.Errorf("expected Kind=ServiceEndpoint, got %q", se.Kind)
	}
	if se.Spec.RedfishEndpointID != req.Spec.RedfishEndpointID {
		t.Errorf("expected Spec.RedfishEndpointID=%q, got %q", req.Spec.RedfishEndpointID, se.Spec.RedfishEndpointID)
	}

	doRequest(t, http.MethodDelete, "/serviceendpoints/"+se.Metadata.UID, nil).Body.Close()
}

// TestGetServiceEndpoints verifies GET /serviceendpoints returns 200 and a non-empty array.
func TestGetServiceEndpoints(t *testing.T) {
	created := createServiceEndpointAndRequire(t, newServiceEndpoint("x3000c0b1", "Chassis"))
	defer func() { doRequest(t, http.MethodDelete, "/serviceendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/serviceendpoints", nil)
	requireStatus(t, resp, http.StatusOK)

	var list []serviceEndpointResponse
	decodeJSON(t, resp, &list)
	if len(list) == 0 {
		t.Error("expected at least one service endpoint in list, got zero")
	}
}

// TestGetServiceEndpoint verifies GET /serviceendpoints/{uid} returns 200 and the correct resource.
func TestGetServiceEndpoint(t *testing.T) {
	created := createServiceEndpointAndRequire(t, newServiceEndpoint("x3000c0b2", "Chassis"))
	defer func() { doRequest(t, http.MethodDelete, "/serviceendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/serviceendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, resp, http.StatusOK)

	var fetched serviceEndpointResponse
	decodeJSON(t, resp, &fetched)
	if fetched.Metadata.UID != created.Metadata.UID {
		t.Errorf("expected UID=%q, got %q", created.Metadata.UID, fetched.Metadata.UID)
	}
}

// TestGetServiceEndpointNotFound verifies GET /serviceendpoints/{uid} for an unknown UID returns 404.
func TestGetServiceEndpointNotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/serviceendpoints/serviceendpoint-does-not-exist", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 for unknown service endpoint, got %d", resp.StatusCode)
	}
}

// TestUpdateServiceEndpoint verifies PUT /serviceendpoints/{uid} returns 200 and persists changes.
func TestUpdateServiceEndpoint(t *testing.T) {
	created := createServiceEndpointAndRequire(t, newServiceEndpoint("x3000c0b3", "Chassis"))
	defer func() { doRequest(t, http.MethodDelete, "/serviceendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	updateReq := updateServiceEndpointRequest{
		Metadata: componentMetadata{Name: "x3000c0b3"},
		Spec: serviceEndpointSpec{
			RedfishEndpointID:   "x3000c0b3",
			RedfishType:         "Manager",
			RedfishEndpointFQDN: "bmc.example.com",
		},
	}
	resp := doRequest(t, http.MethodPut, "/serviceendpoints/"+created.Metadata.UID, updateReq)
	requireStatus(t, resp, http.StatusOK)
	var updated serviceEndpointResponse
	decodeJSON(t, resp, &updated)

	if updated.Spec.RedfishType != "Manager" {
		t.Errorf("expected Spec.RedfishType=Manager after PUT, got %q", updated.Spec.RedfishType)
	}
	if updated.Spec.RedfishEndpointFQDN != "bmc.example.com" {
		t.Errorf("expected Spec.RedfishEndpointFQDN=bmc.example.com after PUT, got %q", updated.Spec.RedfishEndpointFQDN)
	}

	// Confirm via GET
	getResp := doRequest(t, http.MethodGet, "/serviceendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched serviceEndpointResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.RedfishType != "Manager" {
		t.Errorf("GET after PUT: expected Spec.RedfishType=Manager, got %q", fetched.Spec.RedfishType)
	}
}

// TestDeleteServiceEndpoint verifies DELETE /serviceendpoints/{uid} returns 200 and subsequent GET returns 404.
func TestDeleteServiceEndpoint(t *testing.T) {
	created := createServiceEndpointAndRequire(t, newServiceEndpoint("x3000c0b4", "Chassis"))

	delResp := doRequest(t, http.MethodDelete, "/serviceendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.UID != created.Metadata.UID {
		t.Errorf("delete response UID mismatch: want %q, got %q", created.Metadata.UID, delBody.UID)
	}

	getResp := doRequest(t, http.MethodGet, "/serviceendpoints/"+created.Metadata.UID, nil)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", getResp.StatusCode)
	}
}

// TestServiceEndpointLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestServiceEndpointLifecycle(t *testing.T) {
	rfID := "x3000c0b5"

	// POST
	created := createServiceEndpointAndRequire(t, newServiceEndpoint(rfID, "Chassis"))
	uid := created.Metadata.UID
	t.Logf("Created service endpoint UID: %s", uid)

	// GET by UID
	getResp := doRequest(t, http.MethodGet, "/serviceendpoints/"+uid, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched serviceEndpointResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.RedfishEndpointID != rfID {
		t.Errorf("GET: expected Spec.RedfishEndpointID=%q, got %q", rfID, fetched.Spec.RedfishEndpointID)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, "/serviceendpoints/"+uid, updateServiceEndpointRequest{
		Metadata: componentMetadata{Name: rfID},
		Spec:     serviceEndpointSpec{RedfishEndpointID: rfID, RedfishType: "Manager", OdataID: "/redfish/v1/Managers/BMC"},
	})
	requireStatus(t, putResp, http.StatusOK)
	var updated serviceEndpointResponse
	decodeJSON(t, putResp, &updated)
	if updated.Spec.RedfishType != "Manager" {
		t.Errorf("PUT: expected Spec.RedfishType=Manager, got %q", updated.Spec.RedfishType)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, "/serviceendpoints/"+uid, nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	gone := doRequest(t, http.MethodGet, "/serviceendpoints/"+uid, nil)
	defer gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", gone.StatusCode)
	}
}

// TestCreateServiceEndpointDuplicateID verifies that POST /serviceendpoints rejects
// a second resource with the same RedfishURL (resource_id), enforcing uniqueness.
// Note: resource_id for ServiceEndpoints is Spec.RedfishURL (the URL field), so a
// non-empty RedfishURL must be provided for the uniqueness constraint to apply.
func TestCreateServiceEndpointDuplicateID(t *testing.T) {
	const redfishURL = "x3001c0b0-chassis"
	req := createServiceEndpointRequest{
		Metadata: componentMetadata{Name: "x3001c0b0"},
		Spec: serviceEndpointSpec{
			RedfishEndpointID: "x3001c0b0",
			RedfishType:       "Chassis",
			OdataID:           "/redfish/v1/Chassis/x3001c0b0",
			RedfishURL:        redfishURL,
		},
	}
	first := createServiceEndpointAndRequire(t, req)
	defer func() { doRequest(t, http.MethodDelete, "/serviceendpoints/"+first.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodPost, "/serviceendpoints", req)
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate service endpoint URL %q, got HTTP %d", redfishURL, resp.StatusCode)
	}
}
