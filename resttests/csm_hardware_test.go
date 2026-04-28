/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /hsm/v2/Inventory/Hardware routes registered in csm_routes.go.
 * Routes under test:
 *   GET    /hsm/v2/Inventory/Hardware
 *   POST   /hsm/v2/Inventory/Hardware
 *   GET    /hsm/v2/Inventory/Hardware/{id}
 *   PUT    /hsm/v2/Inventory/Hardware/{id}
 *   DELETE /hsm/v2/Inventory/Hardware/{id}
 *
 * Notes on request/response shapes (from csm_hardware_handlers.go):
 *   POST  body   : HardwareArray  { "Hardware": [ <HardwareSpec>, ... ] }
 *   POST  returns: HTTP 201, HardwareArray body
 *   GET / returns: []*HardwareSpec  (flat JSON array)
 *   GET /{id}:    HardwareSpec
 *   PUT  body   : HardwareSpec
 *   PUT  returns: updated HardwareSpec
 *   DELETE /{id}: DeleteResponse { message, uid }
 */

package resttests

import (
	"fmt"
	"net/http"
	"testing"
)

const csmHwBase = "/hsm/v2/Inventory/Hardware"

// ─── CSM hardware request / response shapes ───────────────────────────────────

// csmHardwareArray mirrors cmd/server.HardwareArray.
type csmHardwareArray struct {
	Hardware []*hardwareSpec `json:"Hardware"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// csmHwCreate POSTs a batch of hardware items via the CSM endpoint and asserts 201.
func csmHwCreate(t *testing.T, specs ...*hardwareSpec) {
	t.Helper()
	body := csmHardwareArray{Hardware: specs}
	resp := doRequest(t, http.MethodPost, csmHwBase, body)
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// csmHwGetOne fetches a single hardware item by xname ID and returns the spec.
func csmHwGetOne(t *testing.T, xname string) (*hardwareSpec, int) {
	t.Helper()
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("%s/%s", csmHwBase, xname), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var h hardwareSpec
	decodeJSON(t, resp, &h)
	return &h, resp.StatusCode
}

// csmHwDelete deletes a hardware item by xname ID.
func csmHwDelete(t *testing.T, xname string) {
	t.Helper()
	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmHwBase, xname), nil)
	resp.Body.Close()
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateHardwareCsm verifies that POST /hsm/v2/Inventory/Hardware with a
// HardwareArray body returns HTTP 201.
func TestCreateHardwareCsm(t *testing.T) {
	xname := "x3000c0s6b0n0"
	csmHwCreate(t, &hardwareSpec{ID: xname, Type: "Node"})
	defer csmHwDelete(t, xname)

	// Verify the hardware item exists after creation
	hw, status := csmHwGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GET after POST, got %d", status)
	}
	if hw.ID != xname {
		t.Errorf("expected ID=%q, got %q", xname, hw.ID)
	}
}

// TestCreateHardwareCsmBulk verifies that multiple hardware items can be
// created in a single POST.
func TestCreateHardwareCsmBulk(t *testing.T) {
	xnames := []string{"x3000c0s6b0n1", "x3000c0s6b0n2", "x3000c0s6b0n3"}
	specs := make([]*hardwareSpec, len(xnames))
	for i, x := range xnames {
		specs[i] = &hardwareSpec{ID: x, Type: "Node"}
	}
	csmHwCreate(t, specs...)
	defer func() {
		for _, x := range xnames {
			csmHwDelete(t, x)
		}
	}()

	// Verify each was created
	for _, x := range xnames {
		hw, status := csmHwGetOne(t, x)
		if status != http.StatusOK {
			t.Errorf("expected HTTP 200 for %s, got %d", x, status)
			continue
		}
		if hw.ID != x {
			t.Errorf("expected ID=%q, got %q", x, hw.ID)
		}
	}
}

// TestGetHardwaresCsm verifies that GET /hsm/v2/Inventory/Hardware returns
// HTTP 200 and a list containing the created hardware item.
func TestGetHardwaresCsm(t *testing.T) {
	xname := "x3000c0s6b0n4"
	csmHwCreate(t, &hardwareSpec{ID: xname, Type: "Node"})
	defer csmHwDelete(t, xname)

	resp := doRequest(t, http.MethodGet, csmHwBase, nil)
	requireStatus(t, resp, http.StatusOK)

	var list []*hardwareSpec
	decodeJSON(t, resp, &list)

	found := false
	for _, h := range list {
		if h != nil && h.ID == xname {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("hardware %s not found in GET %s list", xname, csmHwBase)
	}
}

// TestGetHardwareCsm verifies that GET /hsm/v2/Inventory/Hardware/{id} returns
// HTTP 200 and the correct hardware spec.
func TestGetHardwareCsm(t *testing.T) {
	xname := "x3000c0s6b0n5"
	csmHwCreate(t, &hardwareSpec{ID: xname, Type: "Node"})
	defer csmHwDelete(t, xname)

	hw, status := csmHwGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", status)
	}
	if hw.ID != xname {
		t.Errorf("expected ID=%q, got %q", xname, hw.ID)
	}
}

// TestUpdateHardwareCsm verifies that PUT /hsm/v2/Inventory/Hardware/{id}
// updates the hardware spec and returns HTTP 200.
func TestUpdateHardwareCsm(t *testing.T) {
	xname := "x3000c0s6b0n6"
	csmHwCreate(t, &hardwareSpec{ID: xname, Type: "Node"})
	defer csmHwDelete(t, xname)

	updateSpec := hardwareSpec{
		ID:     xname,
		Type:   "Node",
		Status: "Populated",
	}
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmHwBase, xname), updateSpec)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Verify via GET that the update persisted
	hw, status := csmHwGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 after PUT, got %d", status)
	}
	if hw.Status != "Populated" {
		t.Errorf("expected Status=Populated after PUT, got %q", hw.Status)
	}
}

// TestDeleteHardwareCsm verifies that DELETE /hsm/v2/Inventory/Hardware/{id}
// returns HTTP 200 and that a subsequent GET does not return HTTP 200.
func TestDeleteHardwareCsm(t *testing.T) {
	xname := "x3000c0s6b0n7"
	csmHwCreate(t, &hardwareSpec{ID: xname, Type: "Node"})

	// Delete
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmHwBase, xname), nil)
	requireStatus(t, delResp, http.StatusOK)

	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.Message == "" {
		t.Error("expected non-empty message in delete response")
	}

	// Confirm it is gone – the CSM GET handler returns non-200 for missing hardware
	_, status := csmHwGetOne(t, xname)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE %s, got 200", xname)
	}
}

// TestCsmHardwareLifecycle exercises the full POST → GET → PUT → DELETE cycle
// via the CSM /hsm/v2/Inventory/Hardware endpoint.
func TestCsmHardwareLifecycle(t *testing.T) {
	xname := "x3000c0s6b0n8"

	// ── POST ──────────────────────────────────────────────────────────────────
	csmHwCreate(t, &hardwareSpec{ID: xname, Type: "Node"})

	// ── GET by xname ──────────────────────────────────────────────────────────
	hw, status := csmHwGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("POST→GET: expected HTTP 200, got %d", status)
	}
	if hw.ID != xname {
		t.Errorf("POST→GET: expected ID=%q, got %q", xname, hw.ID)
	}

	// ── GET all – hardware must appear ───────────────────────────────────────
	listResp := doRequest(t, http.MethodGet, csmHwBase, nil)
	requireStatus(t, listResp, http.StatusOK)
	var list []*hardwareSpec
	decodeJSON(t, listResp, &list)
	found := false
	for _, h := range list {
		if h != nil && h.ID == xname {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("hardware %s not found in GET list after creation", xname)
	}

	// ── PUT ───────────────────────────────────────────────────────────────────
	putResp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmHwBase, xname), hardwareSpec{
		ID: xname, Type: "Node", Status: "Populated",
	})
	requireStatus(t, putResp, http.StatusOK)
	putResp.Body.Close()

	hw, status = csmHwGetOne(t, xname)
	if status != http.StatusOK {
		t.Fatalf("GET after PUT: expected HTTP 200, got %d", status)
	}
	if hw.Status != "Populated" {
		t.Errorf("PUT: expected Status=Populated, got %q", hw.Status)
	}

	// ── DELETE ────────────────────────────────────────────────────────────────
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmHwBase, xname), nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	_, status = csmHwGetOne(t, xname)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE, still got 200")
	}
}

// TestCreateHardwareCsmDuplicateID verifies that POST /hsm/v2/Inventory/Hardware rejects
// a hardware item whose ID already exists, enforcing resource_id uniqueness.
func TestCreateHardwareCsmDuplicateID(t *testing.T) {
	xname := "x3000c0s7b0n0"
	csmHwCreate(t, &hardwareSpec{ID: xname, Type: "Node"})
	defer csmHwDelete(t, xname)

	resp := doRequest(t, http.MethodPost, csmHwBase, csmHardwareArray{
		Hardware: []*hardwareSpec{{ID: xname, Type: "Node"}},
	})
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate hardware ID %q, got HTTP %d", xname, resp.StatusCode)
	}
}
