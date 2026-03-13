/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /ethernetinterfaces routes registered in routes_generated.go.
 * Routes under test:
 *   GET    /ethernetinterfaces
 *   POST   /ethernetinterfaces
 *   GET    /ethernetinterfaces/{uid}
 *   PUT    /ethernetinterfaces/{uid}
 *   DELETE /ethernetinterfaces/{uid}
 */

package resttests

import (
	"net/http"
	"testing"
)

// ─── Request / response shapes ────────────────────────────────────────────────

type ipAddress struct {
	IPAddress string `json:"IPAddress"`
	Network   string `json:"Network,omitempty"`
}

type ethernetInterfaceSpec struct {
	ID          string      `json:"ID"`
	Description string      `json:"Description,omitempty"`
	MACAddress  string      `json:"MACAddress"`
	LastUpdate  string      `json:"LastUpdate,omitempty"`
	ComponentID string      `json:"ComponentID,omitempty"`
	Type        string      `json:"Type,omitempty"`
	IPAddresses []ipAddress `json:"IPAddresses,omitempty"`
}

type ethernetInterfaceResponse struct {
	APIVersion string                `json:"apiVersion"`
	Kind       string                `json:"kind"`
	Metadata   componentMetadata     `json:"metadata"`
	Spec       ethernetInterfaceSpec `json:"spec"`
}

type createEthernetInterfaceRequest struct {
	Metadata componentMetadata     `json:"metadata"`
	Spec     ethernetInterfaceSpec `json:"spec"`
}

type updateEthernetInterfaceRequest struct {
	Metadata componentMetadata     `json:"metadata,omitempty"`
	Spec     ethernetInterfaceSpec `json:"spec,omitempty"`
}

// newEthernetInterface builds a valid create request.
func newEthernetInterface(id, mac string) createEthernetInterfaceRequest {
	return createEthernetInterfaceRequest{
		Metadata: componentMetadata{Name: id},
		Spec: ethernetInterfaceSpec{
			ID:          id,
			MACAddress:  mac,
			ComponentID: "x3000c0s0b0n0",
			Type:        "HostLAN",
			IPAddresses: []ipAddress{{IPAddress: "10.0.0.1"}},
		},
	}
}

// createEthernetInterfaceAndRequire POSTs an ethernet interface and validates 201.
func createEthernetInterfaceAndRequire(t *testing.T, req createEthernetInterfaceRequest) ethernetInterfaceResponse {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/ethernetinterfaces", req)
	requireStatus(t, resp, http.StatusCreated)
	var created ethernetInterfaceResponse
	decodeJSON(t, resp, &created)
	if created.Metadata.UID == "" {
		t.Fatal("expected non-empty UID in created ethernet interface response")
	}
	return created
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateEthernetInterface verifies POST /ethernetinterfaces returns 201 with a generated UID.
func TestCreateEthernetInterface(t *testing.T) {
	req := newEthernetInterface("a0:00:00:00:00:01", "a0:00:00:00:00:01")
	ei := createEthernetInterfaceAndRequire(t, req)

	if ei.Kind != "EthernetInterface" {
		t.Errorf("expected Kind=EthernetInterface, got %q", ei.Kind)
	}
	if ei.Spec.MACAddress != req.Spec.MACAddress {
		t.Errorf("expected Spec.MACAddress=%q, got %q", req.Spec.MACAddress, ei.Spec.MACAddress)
	}

	doRequest(t, http.MethodDelete, "/ethernetinterfaces/"+ei.Metadata.UID, nil).Body.Close()
}

// TestGetEthernetInterfaces verifies GET /ethernetinterfaces returns 200 and a non-empty list.
func TestGetEthernetInterfaces(t *testing.T) {
	created := createEthernetInterfaceAndRequire(t, newEthernetInterface("a0:00:00:00:00:02", "a0:00:00:00:00:02"))
	defer func() { doRequest(t, http.MethodDelete, "/ethernetinterfaces/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/ethernetinterfaces", nil)
	requireStatus(t, resp, http.StatusOK)

	var list []ethernetInterfaceResponse
	decodeJSON(t, resp, &list)
	if len(list) == 0 {
		t.Error("expected at least one ethernet interface in list, got zero")
	}
}

// TestGetEthernetInterface verifies GET /ethernetinterfaces/{uid} returns 200 and the correct resource.
func TestGetEthernetInterface(t *testing.T) {
	created := createEthernetInterfaceAndRequire(t, newEthernetInterface("a0:00:00:00:00:03", "a0:00:00:00:00:03"))
	defer func() { doRequest(t, http.MethodDelete, "/ethernetinterfaces/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/ethernetinterfaces/"+created.Metadata.UID, nil)
	requireStatus(t, resp, http.StatusOK)

	var fetched ethernetInterfaceResponse
	decodeJSON(t, resp, &fetched)
	if fetched.Metadata.UID != created.Metadata.UID {
		t.Errorf("expected UID=%q, got %q", created.Metadata.UID, fetched.Metadata.UID)
	}
	if fetched.Spec.MACAddress != created.Spec.MACAddress {
		t.Errorf("expected Spec.MACAddress=%q, got %q", created.Spec.MACAddress, fetched.Spec.MACAddress)
	}
}

// TestGetEthernetInterfaceNotFound verifies GET /ethernetinterfaces/{uid} for unknown UID returns 404.
func TestGetEthernetInterfaceNotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/ethernetinterfaces/ethernetinterface-does-not-exist", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 for unknown ethernet interface, got %d", resp.StatusCode)
	}
}

// TestUpdateEthernetInterface verifies PUT /ethernetinterfaces/{uid} returns 200 and persists changes.
func TestUpdateEthernetInterface(t *testing.T) {
	created := createEthernetInterfaceAndRequire(t, newEthernetInterface("a0:00:00:00:00:04", "a0:00:00:00:00:04"))
	defer func() { doRequest(t, http.MethodDelete, "/ethernetinterfaces/"+created.Metadata.UID, nil).Body.Close() }()

	updateReq := updateEthernetInterfaceRequest{
		Metadata: componentMetadata{Name: "a0:00:00:00:00:04"},
		Spec: ethernetInterfaceSpec{
			ID:          "a0:00:00:00:00:04",
			MACAddress:  "a0:00:00:00:00:04",
			ComponentID: "x3000c0s0b0n0",
			Type:        "HostLAN",
			IPAddresses: []ipAddress{{IPAddress: "10.0.0.99"}},
		},
	}
	resp := doRequest(t, http.MethodPut, "/ethernetinterfaces/"+created.Metadata.UID, updateReq)
	requireStatus(t, resp, http.StatusOK)
	var updated ethernetInterfaceResponse
	decodeJSON(t, resp, &updated)

	if len(updated.Spec.IPAddresses) == 0 || updated.Spec.IPAddresses[0].IPAddress != "10.0.0.99" {
		t.Errorf("expected Spec.IPAddresses[0]=10.0.0.99 after PUT")
	}

	// Confirm via GET
	getResp := doRequest(t, http.MethodGet, "/ethernetinterfaces/"+created.Metadata.UID, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched ethernetInterfaceResponse
	decodeJSON(t, getResp, &fetched)
	if len(fetched.Spec.IPAddresses) == 0 || fetched.Spec.IPAddresses[0].IPAddress != "10.0.0.99" {
		t.Errorf("GET after PUT: expected Spec.IPAddresses[0]=10.0.0.99")
	}
}

// TestDeleteEthernetInterface verifies DELETE /ethernetinterfaces/{uid} returns 200 and subsequent GET returns 404.
func TestDeleteEthernetInterface(t *testing.T) {
	created := createEthernetInterfaceAndRequire(t, newEthernetInterface("a0:00:00:00:00:05", "a0:00:00:00:00:05"))

	delResp := doRequest(t, http.MethodDelete, "/ethernetinterfaces/"+created.Metadata.UID, nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.UID != created.Metadata.UID {
		t.Errorf("delete response UID mismatch: want %q, got %q", created.Metadata.UID, delBody.UID)
	}

	getResp := doRequest(t, http.MethodGet, "/ethernetinterfaces/"+created.Metadata.UID, nil)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", getResp.StatusCode)
	}
}

// TestEthernetInterfaceLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestEthernetInterfaceLifecycle(t *testing.T) {
	mac := "a0:00:00:00:00:06"

	// POST
	created := createEthernetInterfaceAndRequire(t, newEthernetInterface(mac, mac))
	uid := created.Metadata.UID
	t.Logf("Created ethernet interface UID: %s", uid)

	// GET by UID
	getResp := doRequest(t, http.MethodGet, "/ethernetinterfaces/"+uid, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched ethernetInterfaceResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.MACAddress != mac {
		t.Errorf("GET: expected Spec.MACAddress=%q, got %q", mac, fetched.Spec.MACAddress)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, "/ethernetinterfaces/"+uid, updateEthernetInterfaceRequest{
		Metadata: componentMetadata{Name: mac},
		Spec: ethernetInterfaceSpec{
			ID: mac, MACAddress: mac, ComponentID: "x3000c0s0b0n1", Type: "HostLAN",
			IPAddresses: []ipAddress{{IPAddress: "10.1.2.3"}},
		},
	})
	requireStatus(t, putResp, http.StatusOK)
	var updated ethernetInterfaceResponse
	decodeJSON(t, putResp, &updated)
	if updated.Spec.ComponentID != "x3000c0s0b0n1" {
		t.Errorf("PUT: expected Spec.ComponentID=x3000c0s0b0n1, got %q", updated.Spec.ComponentID)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, "/ethernetinterfaces/"+uid, nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	gone := doRequest(t, http.MethodGet, "/ethernetinterfaces/"+uid, nil)
	defer gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", gone.StatusCode)
	}
}

// TestCreateEthernetInterfaceDuplicateID verifies that POST /ethernetinterfaces rejects
// a second resource with the same Spec.ID, enforcing resource_id uniqueness.
func TestCreateEthernetInterfaceDuplicateID(t *testing.T) {
	mac := "c0:00:00:00:00:01"
	first := createEthernetInterfaceAndRequire(t, newEthernetInterface(mac, mac))
	defer func() { doRequest(t, http.MethodDelete, "/ethernetinterfaces/"+first.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodPost, "/ethernetinterfaces", newEthernetInterface(mac, mac))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate ethernet interface ID %q, got HTTP %d", mac, resp.StatusCode)
	}
}
