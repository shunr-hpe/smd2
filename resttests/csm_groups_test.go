/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /hsm/v2/groups routes registered in csm_routes.go.
 * Routes under test:
 *   GET    /hsm/v2/groups
 *   POST   /hsm/v2/groups
 *   GET    /hsm/v2/groups/{group_label}
 *   PUT    /hsm/v2/groups/{group_label}
 *   DELETE /hsm/v2/groups/{group_label}
 *
 * Notes on request/response shapes (from csm_groups.go):
 *   POST  body   : GroupSpec (single object)
 *   POST  returns: HTTP 201, echoes GroupSpec
 *   GET / returns: []*GroupSpec  (plain JSON array)
 *   GET /{label} : GroupSpec (spec only)
 *   PUT  body   : GroupSpec
 *   PUT  returns: GroupSpec (updated)
 *   DELETE /{label}: DeleteResponse { message, uid }
 *   ID key       : GroupSpec.label
 */

package resttests

import (
	"fmt"
	"net/http"
	"testing"
)

const csmGroupBase = "/hsm/v2/groups"

// ─── CSM request / response shapes ───────────────────────────────────────────

type csmGroupMembers struct {
	IDs []string `json:"ids"`
}

type csmGroupSpec struct {
	Label          string          `json:"label"`
	Description    string          `json:"description,omitempty"`
	ExclusiveGroup string          `json:"exclusiveGroup,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
	Members        csmGroupMembers `json:"members"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// csmGroupCreate POSTs a single GroupSpec via the CSM endpoint and asserts 201.
func csmGroupCreate(t *testing.T, spec csmGroupSpec) {
	t.Helper()
	resp := doRequest(t, http.MethodPost, csmGroupBase, spec)
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// csmGroupGetOne fetches a single group by label.
func csmGroupGetOne(t *testing.T, label string) (*csmGroupSpec, int) {
	t.Helper()
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("%s/%s", csmGroupBase, label), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var spec csmGroupSpec
	decodeJSON(t, resp, &spec)
	return &spec, resp.StatusCode
}

// csmGroupDelete deletes a group by label.
func csmGroupDelete(t *testing.T, label string) {
	t.Helper()
	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmGroupBase, label), nil)
	resp.Body.Close()
}

func newCsmGroup(label string, memberIDs ...string) csmGroupSpec {
	ids := memberIDs
	if ids == nil {
		ids = []string{}
	}
	return csmGroupSpec{
		Label:   label,
		Members: csmGroupMembers{IDs: ids},
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateGroupCsm verifies POST /hsm/v2/groups returns 201.
func TestCreateGroupCsm(t *testing.T) {
	label := "csm-compute-nodes"
	csmGroupCreate(t, newCsmGroup(label, "x3000c0s0b0n0", "x3000c0s0b0n1"))
	defer csmGroupDelete(t, label)

	spec, status := csmGroupGetOne(t, label)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 for GET after POST, got %d", status)
	}
	if spec.Label != label {
		t.Errorf("expected Label=%q, got %q", label, spec.Label)
	}
	if len(spec.Members.IDs) != 2 {
		t.Errorf("expected 2 member IDs, got %d", len(spec.Members.IDs))
	}
}

// TestGetGroupsCsm verifies GET /hsm/v2/groups returns a plain array
// containing the created group.
func TestGetGroupsCsm(t *testing.T) {
	label := "csm-io-nodes"
	csmGroupCreate(t, newCsmGroup(label))
	defer csmGroupDelete(t, label)

	resp := doRequest(t, http.MethodGet, csmGroupBase, nil)
	requireStatus(t, resp, http.StatusOK)

	var list []csmGroupSpec
	decodeJSON(t, resp, &list)

	found := false
	for _, g := range list {
		if g.Label == label {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("group %q not found in GET %s list", label, csmGroupBase)
	}
}

// TestGetGroupCsm verifies GET /hsm/v2/groups/{group_label} returns 200
// and the correct GroupSpec.
func TestGetGroupCsm(t *testing.T) {
	label := "csm-storage-nodes"
	csmGroupCreate(t, newCsmGroup(label, "x3000c0s1b0n0"))
	defer csmGroupDelete(t, label)

	spec, status := csmGroupGetOne(t, label)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", status)
	}
	if spec.Label != label {
		t.Errorf("expected Label=%q, got %q", label, spec.Label)
	}
}

// TestUpdateGroupCsm verifies PUT /hsm/v2/groups/{group_label} updates
// the GroupSpec and returns 200.
func TestUpdateGroupCsm(t *testing.T) {
	label := "csm-mgmt-nodes"
	csmGroupCreate(t, newCsmGroup(label, "x3000c0s0b0n0"))
	defer csmGroupDelete(t, label)

	updateSpec := csmGroupSpec{
		Label:       label,
		Description: "Updated management nodes",
		Members:     csmGroupMembers{IDs: []string{"x3000c0s0b0n0", "x3000c0s0b0n1", "x3000c0s0b0n2"}},
	}
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmGroupBase, label), updateSpec)
	requireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	spec, status := csmGroupGetOne(t, label)
	if status != http.StatusOK {
		t.Fatalf("expected HTTP 200 after PUT, got %d", status)
	}
	if spec.Description != "Updated management nodes" {
		t.Errorf("expected Description=Updated management nodes after PUT, got %q", spec.Description)
	}
	if len(spec.Members.IDs) != 3 {
		t.Errorf("expected 3 members after PUT, got %d", len(spec.Members.IDs))
	}
}

// TestDeleteGroupCsm verifies DELETE /hsm/v2/groups/{group_label} returns 200
// and that a subsequent GET does not return 200.
func TestDeleteGroupCsm(t *testing.T) {
	label := "csm-ephemeral-group"
	csmGroupCreate(t, newCsmGroup(label))

	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmGroupBase, label), nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.Message == "" {
		t.Error("expected non-empty message in delete response")
	}

	_, status := csmGroupGetOne(t, label)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE %q, still got 200", label)
	}
}

// TestCsmGroupLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestCsmGroupLifecycle(t *testing.T) {
	label := "csm-lifecycle-group"

	// POST
	csmGroupCreate(t, newCsmGroup(label, "x3000c0s0b0n0"))

	// GET
	spec, status := csmGroupGetOne(t, label)
	if status != http.StatusOK {
		t.Fatalf("POST→GET: expected HTTP 200, got %d", status)
	}
	if spec.Label != label {
		t.Errorf("POST→GET: expected Label=%q, got %q", label, spec.Label)
	}

	// GET all – must appear
	listResp := doRequest(t, http.MethodGet, csmGroupBase, nil)
	requireStatus(t, listResp, http.StatusOK)
	var list []csmGroupSpec
	decodeJSON(t, listResp, &list)
	found := false
	for _, g := range list {
		if g.Label == label {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("group %q not found in GET list after creation", label)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/%s", csmGroupBase, label), csmGroupSpec{
		Label:       label,
		Description: "Lifecycle updated",
		Tags:        []string{"compute", "test"},
		Members:     csmGroupMembers{IDs: []string{"x3000c0s0b0n0", "x3000c0s0b0n1"}},
	})
	requireStatus(t, putResp, http.StatusOK)
	putResp.Body.Close()

	spec, status = csmGroupGetOne(t, label)
	if status != http.StatusOK {
		t.Fatalf("GET after PUT: expected HTTP 200, got %d", status)
	}
	if spec.Description != "Lifecycle updated" {
		t.Errorf("PUT: expected Description=Lifecycle updated, got %q", spec.Description)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/%s", csmGroupBase, label), nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	_, status = csmGroupGetOne(t, label)
	if status == http.StatusOK {
		t.Errorf("expected non-200 after DELETE, still got 200")
	}
}

// TestCreateGroupCsmDuplicateID verifies that POST /hsm/v2/groups rejects a group
// whose label already exists, enforcing resource_id uniqueness.
func TestCreateGroupCsmDuplicateID(t *testing.T) {
	label := "csm-duplicate-test-group"
	csmGroupCreate(t, newCsmGroup(label))
	defer csmGroupDelete(t, label)

	resp := doRequest(t, http.MethodPost, csmGroupBase, newCsmGroup(label))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate group label %q, got HTTP %d", label, resp.StatusCode)
	}
}
