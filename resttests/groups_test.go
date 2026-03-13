/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests for the /groups routes registered in routes_generated.go.
 * Routes under test:
 *   GET    /groups
 *   POST   /groups
 *   GET    /groups/{uid}
 *   PUT    /groups/{uid}
 *   DELETE /groups/{uid}
 */

package resttests

import (
	"net/http"
	"testing"
)

// ─── Request / response shapes ────────────────────────────────────────────────

type groupMembers struct {
	IDs []string `json:"ids"`
}

type groupSpec struct {
	Label          string       `json:"label"`
	Description    string       `json:"description,omitempty"`
	ExclusiveGroup string       `json:"exclusiveGroup,omitempty"`
	Tags           []string     `json:"tags,omitempty"`
	Members        groupMembers `json:"members"`
}

type groupResponse struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   componentMetadata `json:"metadata"`
	Spec       groupSpec         `json:"spec"`
}

type createGroupRequest struct {
	Metadata componentMetadata `json:"metadata"`
	Spec     groupSpec         `json:"spec"`
}

type updateGroupRequest struct {
	Metadata componentMetadata `json:"metadata,omitempty"`
	Spec     groupSpec         `json:"spec,omitempty"`
}

// newGroup builds a valid create request.
func newGroup(label string, memberIDs ...string) createGroupRequest {
	ids := memberIDs
	if ids == nil {
		ids = []string{}
	}
	return createGroupRequest{
		Metadata: componentMetadata{Name: label},
		Spec: groupSpec{
			Label:   label,
			Members: groupMembers{IDs: ids},
		},
	}
}

// createGroupAndRequire POSTs a group and validates 201.
func createGroupAndRequire(t *testing.T, req createGroupRequest) groupResponse {
	t.Helper()
	resp := doRequest(t, http.MethodPost, "/groups", req)
	requireStatus(t, resp, http.StatusCreated)
	var created groupResponse
	decodeJSON(t, resp, &created)
	if created.Metadata.UID == "" {
		t.Fatal("expected non-empty UID in created group response")
	}
	return created
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestCreateGroup verifies POST /groups returns 201 with a generated UID.
func TestCreateGroup(t *testing.T) {
	req := newGroup("compute-nodes", "x3000c0s0b0n0", "x3000c0s0b0n1")
	g := createGroupAndRequire(t, req)

	if g.Kind != "Group" {
		t.Errorf("expected Kind=Group, got %q", g.Kind)
	}
	if g.Spec.Label != req.Spec.Label {
		t.Errorf("expected Spec.Label=%q, got %q", req.Spec.Label, g.Spec.Label)
	}

	doRequest(t, http.MethodDelete, "/groups/"+g.Metadata.UID, nil).Body.Close()
}

// TestGetGroups verifies GET /groups returns 200 and a non-empty list.
func TestGetGroups(t *testing.T) {
	created := createGroupAndRequire(t, newGroup("io-nodes"))
	defer func() { doRequest(t, http.MethodDelete, "/groups/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/groups", nil)
	requireStatus(t, resp, http.StatusOK)

	var list []groupResponse
	decodeJSON(t, resp, &list)
	if len(list) == 0 {
		t.Error("expected at least one group in list, got zero")
	}
}

// TestGetGroup verifies GET /groups/{uid} returns 200 and the correct resource.
func TestGetGroup(t *testing.T) {
	created := createGroupAndRequire(t, newGroup("storage-nodes"))
	defer func() { doRequest(t, http.MethodDelete, "/groups/"+created.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodGet, "/groups/"+created.Metadata.UID, nil)
	requireStatus(t, resp, http.StatusOK)

	var fetched groupResponse
	decodeJSON(t, resp, &fetched)
	if fetched.Metadata.UID != created.Metadata.UID {
		t.Errorf("expected UID=%q, got %q", created.Metadata.UID, fetched.Metadata.UID)
	}
	if fetched.Spec.Label != created.Spec.Label {
		t.Errorf("expected Spec.Label=%q, got %q", created.Spec.Label, fetched.Spec.Label)
	}
}

// TestGetGroupNotFound verifies GET /groups/{uid} for an unknown UID returns 404.
func TestGetGroupNotFound(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/groups/group-does-not-exist", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 for unknown group, got %d", resp.StatusCode)
	}
}

// TestUpdateGroup verifies PUT /groups/{uid} returns 200 and persists changes.
func TestUpdateGroup(t *testing.T) {
	created := createGroupAndRequire(t, newGroup("mgmt-nodes", "x3000c0s0b0n0"))
	defer func() { doRequest(t, http.MethodDelete, "/groups/"+created.Metadata.UID, nil).Body.Close() }()

	updateReq := updateGroupRequest{
		Metadata: componentMetadata{Name: "mgmt-nodes"},
		Spec: groupSpec{
			Label:       "mgmt-nodes",
			Description: "Management nodes group",
			Members:     groupMembers{IDs: []string{"x3000c0s0b0n0", "x3000c0s0b0n1"}},
		},
	}
	resp := doRequest(t, http.MethodPut, "/groups/"+created.Metadata.UID, updateReq)
	requireStatus(t, resp, http.StatusOK)
	var updated groupResponse
	decodeJSON(t, resp, &updated)

	if updated.Spec.Description != "Management nodes group" {
		t.Errorf("expected Spec.Description=Management nodes group after PUT, got %q", updated.Spec.Description)
	}
	if len(updated.Spec.Members.IDs) != 2 {
		t.Errorf("expected 2 members after PUT, got %d", len(updated.Spec.Members.IDs))
	}

	// Confirm via GET
	getResp := doRequest(t, http.MethodGet, "/groups/"+created.Metadata.UID, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched groupResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.Description != "Management nodes group" {
		t.Errorf("GET after PUT: expected Spec.Description=Management nodes group, got %q", fetched.Spec.Description)
	}
}

// TestDeleteGroup verifies DELETE /groups/{uid} returns 200 and subsequent GET returns 404.
func TestDeleteGroup(t *testing.T) {
	created := createGroupAndRequire(t, newGroup("ephemeral-group"))

	delResp := doRequest(t, http.MethodDelete, "/groups/"+created.Metadata.UID, nil)
	requireStatus(t, delResp, http.StatusOK)
	var delBody deleteComponentResponse
	decodeJSON(t, delResp, &delBody)
	if delBody.UID != created.Metadata.UID {
		t.Errorf("delete response UID mismatch: want %q, got %q", created.Metadata.UID, delBody.UID)
	}

	getResp := doRequest(t, http.MethodGet, "/groups/"+created.Metadata.UID, nil)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", getResp.StatusCode)
	}
}

// TestGroupLifecycle exercises the full POST → GET → PUT → DELETE cycle.
func TestGroupLifecycle(t *testing.T) {
	label := "lifecycle-test-group"

	// POST
	created := createGroupAndRequire(t, newGroup(label, "x3000c0s0b0n0"))
	uid := created.Metadata.UID
	t.Logf("Created group UID: %s", uid)

	// GET by UID
	getResp := doRequest(t, http.MethodGet, "/groups/"+uid, nil)
	requireStatus(t, getResp, http.StatusOK)
	var fetched groupResponse
	decodeJSON(t, getResp, &fetched)
	if fetched.Spec.Label != label {
		t.Errorf("GET: expected Spec.Label=%q, got %q", label, fetched.Spec.Label)
	}

	// GET list – group must appear
	listResp := doRequest(t, http.MethodGet, "/groups", nil)
	requireStatus(t, listResp, http.StatusOK)
	var list []groupResponse
	decodeJSON(t, listResp, &list)
	found := false
	for _, g := range list {
		if g.Metadata.UID == uid {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("group %s not found in GET /groups list", uid)
	}

	// PUT
	putResp := doRequest(t, http.MethodPut, "/groups/"+uid, updateGroupRequest{
		Metadata: componentMetadata{Name: label},
		Spec: groupSpec{
			Label:       label,
			Description: "Updated description",
			Tags:        []string{"compute", "test"},
			Members:     groupMembers{IDs: []string{"x3000c0s0b0n0", "x3000c0s0b0n1", "x3000c0s0b0n2"}},
		},
	})
	requireStatus(t, putResp, http.StatusOK)
	var updated groupResponse
	decodeJSON(t, putResp, &updated)
	if updated.Spec.Description != "Updated description" {
		t.Errorf("PUT: expected Spec.Description=Updated description, got %q", updated.Spec.Description)
	}

	// DELETE
	delResp := doRequest(t, http.MethodDelete, "/groups/"+uid, nil)
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	gone := doRequest(t, http.MethodGet, "/groups/"+uid, nil)
	defer gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after DELETE, got %d", gone.StatusCode)
	}
}

// TestCreateGroupDuplicateID verifies that POST /groups rejects a second resource
// with the same Spec.Label, enforcing resource_id uniqueness.
func TestCreateGroupDuplicateID(t *testing.T) {
	label := "duplicate-test-group"
	first := createGroupAndRequire(t, newGroup(label))
	defer func() { doRequest(t, http.MethodDelete, "/groups/"+first.Metadata.UID, nil).Body.Close() }()

	resp := doRequest(t, http.MethodPost, "/groups", newGroup(label))
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected non-2xx on duplicate group label %q, got HTTP %d", label, resp.StatusCode)
	}
}
