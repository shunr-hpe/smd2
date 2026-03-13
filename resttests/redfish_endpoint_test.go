/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /redfishendpoints routes registered in routes_generated.go.
 * Routes under test:
 *   GET    /redfishendpoints
 *   POST   /redfishendpoints
 *   GET    /redfishendpoints/{uid}
 *   PUT    /redfishendpoints/{uid}
 *   DELETE /redfishendpoints/{uid}
 */

package resttests

import (
	"net/http"
	"testing"
)

// ─── Request / response shapes ────────────────────────────────────────────────

type discoveryInfo struct {
	LastDiscoveryAttempt string `json:"LastDiscoveryAttempt,omitempty"`
	LastDiscoveryStatus  string `json:"LastDiscoveryStatus"`
	RedfishVersion       string `json:"RedfishVersion,omitempty"`
}

type redfishEndpointSpec struct {
	ID                 string        `json:"ID"`
	Type               string        `json:"Type"`
	Hostname           string        `json:"Hostname,omitempty"`
	Domain             string        `json:"Domain,omitempty"`
	FQDN               string        `json:"FQDN,omitempty"`
	Enabled            bool          `json:"Enabled,omitempty"`
	User               string        `json:"User,omitempty"`
	Password           string        `json:"Password,omitempty"`
	RediscoverOnUpdate bool          `json:"RediscoverOnUpdate,omitempty"`
	DiscoveryInfo      discoveryInfo `json:"DiscoveryInfo,omitempty"`
}

type redfishEndpointResponse struct {
	APIVersion string              `json:"apiVersion"`
	Kind       string              `json:"kind"`
	Metadata   componentMetadata   `json:"metadata"`
	Spec       redfishEndpointSpec `json:"spec"`
}

type createRedfishEndpointRequest struct {
	Metadata componentMetadata   `json:"metadata"`
	Spec     redfishEndpointSpec `json:"spec"`
}

type updateRedfishEndpointRequest struct {
	Metadata componentMetadata   `json:"metadata,omitempty"`
	Spec     redfishEndpointSpec `json:"spec,omitempty"`
}

// newRedfishEndpoint builds a valid create request.
func newRedfishEndpoint(id, hostname string) createRedfishEndpointRequest {
	return createRedfishEndpointRequest{
		Metadata: componentMetadata{Name: id},
		Spec: redfishEndpointSpec{
			ID:       id,
			Type:     "NodeBMC",
			Hostname: hostname,
			FQDN:     hostname,
			Enabled:  true,
			User:     "root",
			Password: "password",
			DiscoveryInfo: discoveryInfo{
				LastDiscoveryStatus: "NotYetQueried",
			},
		},
	}
}

// createRedfishEndpointAndRequire POSTs a redfish endpoint and validates 201.
func createRedfishEndpointAndRequire(t *testing.T, req createRedfishEndpointRequest) redfishEndpointResponse {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/redfishendpoints", req)
	requireStatus(t, resp, http.StatusCreated)
	var created redfishEndpointResponse
	decodeJSON(t, resp, &created)
	if created.Metadata.UID == "" {
		t.Fatal("expected non-empty UID in created redfish endpoint response")
	}
	return created
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateRedfishEndpoint verifies POST /redfishendpoints returns 201 with a generated UID.
func TestCreateRedfishEndpoint(t *testing.T) {
	req := newRedfishEndpoint("x3000c0s1b0", "bmc1.example.com")
	re := createRedfishEndpointAndRequire(t, req)

	if re.Kind != "RedfishEndpoint" {
		t.Errorf("expected Kind=RedfishEndpoint, got %q", re.Kind)
	}
	if re.Spec.ID != req.Spec.ID {
		t.Errorf("expected Spec.ID=%q, got %q", req.Spec.ID, re.Spec.ID)
	}

	doRequest(t, http.MethodDelete, "/redfishendpoints/"+re.Metadata.UID, nil).Body.Close()
}

// TestGetRedfishEndpoints verifies GET /redfishendpoints returns 200 and a non-empty array.
func TestGetRedfishEndpoints(t *testing.T) {
	created := createRedfishEndpointAndRequire(t, newRedfishEndpoint("x3000c0s2b0", "bmc2.example.com"))
	defer func() { doRequest(t, http.MethodDelete, "/redfishendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/redfishendpoints", nil)
	requireStatus(t, resp, http.StatusOK)

	var list []redfishEndpointResponse
	decodeJSON(t, resp, &list)
	if len(list) == 0 {
		t.Error("expected at least one redfish endpoint in list, got zero")
	}
}

// TestGetRedfishEndpoint verifies GET /redfishendpoints/{uid} returns 200 and the correct resource.
func TestGetRedfishEndpoint(t *testing.T) {
	created := createRedfishEndpointAndRequire(t, newRedfishEndpoint("x3000c0s3b0", "bmc3.example.com"))
	defer func() { doRequest(t, http.MethodDelete, "/redfishendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/redfishendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, resp, http.StatusOK)

	var fetched redfishEndpointResponse
	decodeJSON(t, resp, &fetched)
	if fetched.Metadata.UID != created.Metadata.UID {
		t.Errorf("expected UID=%q, got %q", created.Metadata.UID, fetched.Metadata.UID)
	}
	if fetched.Spec.ID != created.Spec.ID {
		t.Errorf("expected Spec.ID=%q, got %q", created.Spec.ID, fetched.Spec.ID)
	}
}

// TestGetRedfishEndpointNotFound verifies GET /redfishendpoints/{uid} for unknown UID returns 404.
func TestGetRedfishEndpointNotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/redfishendpoints/redfishendpoint-does-not-exist", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 for unknown redfish endpoint, got %d", resp.StatusCode)
	}
}

// TestUpdateRedfishEndpoint verifies PUT /redfishendpoints/{uid} returns 200 and persists changes.
func TestUpdateRedfishEndpoint(t *testing.T) {
	created := createRedfishEndpointAndRequire(t, newRedfishEndpoint("x3000c0s4b0", "bmc4.example.com"))
	defer func() { doRequest(t, http.MethodDelete, "/redfishendpoints/"+created.Metadata.UID, nil).Body.Close() }()

	updateReq := updateRedfishEndpointRequest{
		Metadata: componentMetadata{Name: "x3000c0s4b0"},
		Spec: redfishEndpointSpec{
			ID:       "x3000c0s4b0",
			Type:     "NodeBMC",
			Hostname: "bmc4-updated.example.com",
			FQDN:     "bmc4-updated.example.com",
			Enabled:  true,
			User:     "admin",
			Password: "newpassword",
			DiscoveryInfo: discoveryInfo{
				LastDiscoveryStatus: "DiscoverOK",
			},
		},
	}
	resp := doRequest(t, http.MethodPut, "/redfishendpoints/"+created.Metadata.UID, updateReq)
	requireStatus(t, resp, http.StatusOK)
	var updated redfishEndpointResponse
	decodeJSON(t, resp, &updated)

	if updated.Spec.Hostname != "bmc4-updated.example.com" {
		t.Errorf("expected Spec.Hostname=bmc4-updated.example.com after PUT, got %q", updated.Spec.Hostname)
	}
	if updated.Spec.User != "admin" {
		t.Errorf("expected Spec.User=admin after PUT, got %q", updated.Spec.User)
	}

	// Confirm via GET
	getResp := doRequest(t, http.MethodGet, "/redfishendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched redfishEndpointResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.Hostname != "bmc4-updated.example.com" {
		t.Errorf("GET after PUT: expected Spec.Hostname=bmc4-updated.example.com, got %q", fetched.Spec.Hostname)
	}
}

// TestDeleteRedfishEndpoint verifies DELETE /redfishendpoints/{uid} returns 200 and subsequent GET returns 404.
func TestDeleteRedfishEndpoint(t *testing.T) {
	created := createRedfishEndpointAndRequire(t, newRedfishEndpoint("x3000c0s5b0", "bmc5.example.com"))

	delResp := doRequest(t, http.MethodDelete, "/redfishendpoints/"+created.Metadata.UID, nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.UID != created.Metadata.UID {
		t.Errorf("delete response UID mismatch: want %q, got %q", created.Metadata.UID, delBody.UID)
	}

	getResp := doRequest(t, http.MethodGet, "/redfishendpoints/"+created.Metadata.UID, nil)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", getResp.StatusCode)
	}
}

// TestRedfishEndpointLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestRedfishEndpointLifecycle(t *testing.T) {
	id := "x3000c0s6b0"

	// POST
	created := createRedfishEndpointAndRequire(t, newRedfishEndpoint(id, "bmc6.example.com"))
	uid := created.Metadata.UID
	t.Logf("Created redfish endpoint UID: %s", uid)

	// GET by UID
	getResp := doRequest(t, http.MethodGet, "/redfishendpoints/"+uid, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched redfishEndpointResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.ID != id {
		t.Errorf("GET: expected Spec.ID=%q, got %q", id, fetched.Spec.ID)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, "/redfishendpoints/"+uid, updateRedfishEndpointRequest{
		Metadata: componentMetadata{Name: id},
		Spec: redfishEndpointSpec{
			ID: id, Type: "NodeBMC", Hostname: "bmc6-new.example.com", FQDN: "bmc6-new.example.com",
			Enabled: true, User: "root", Password: "pass",
			DiscoveryInfo: discoveryInfo{LastDiscoveryStatus: "DiscoverOK"},
		},
	})
	requireStatus(t, putResp, http.StatusOK)
	var updated redfishEndpointResponse
	decodeJSON(t, putResp, &updated)
	if updated.Spec.Hostname != "bmc6-new.example.com" {
		t.Errorf("PUT: expected Spec.Hostname=bmc6-new.example.com, got %q", updated.Spec.Hostname)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, "/redfishendpoints/"+uid, nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	gone := doRequest(t, http.MethodGet, "/redfishendpoints/"+uid, nil)
	defer gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", gone.StatusCode)
	}
}

// TestCreateRedfishEndpointDuplicateID verifies that POST /redfishendpoints rejects
// a second resource with the same Spec.ID, enforcing resource_id uniqueness.
func TestCreateRedfishEndpointDuplicateID(t *testing.T) {
	id := "x3001c0s0b0"
	first := createRedfishEndpointAndRequire(t, newRedfishEndpoint(id, "bmc-dup.example.com"))
	defer func() { doRequest(t, http.MethodDelete, "/redfishendpoints/"+first.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodPost, "/redfishendpoints", newRedfishEndpoint(id, "bmc-dup.example.com"))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate redfish endpoint ID %q, got HTTP %d", id, resp.StatusCode)
	}
}
