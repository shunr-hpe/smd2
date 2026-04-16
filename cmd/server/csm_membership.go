// Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"net/http"

	v1 "github.com/OpenCHAMI/inventory-service/apis/inventory-service.openchami.org/v1"
	"github.com/OpenCHAMI/inventory-service/cmd/plugins"
	"github.com/go-chi/chi/v5"
)

func GetMembershipsCsm(w http.ResponseWriter, r *http.Request) {
	groups, err := plugins.Store.LoadAllGroups(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load groups: %w", err))
		return
	}

	componentEndpoints, err := plugins.Store.LoadAllComponentEndpoints(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load componentendpoints: %w", err))
		return
	}

	m := componentsToEmptyMemberships(componentEndpoints)
	m = groupsToMemberships(groups, m)
	memberships := make([]*v1.Membership, 0)
	for _, v := range m {
		memberships = append(memberships, v)
	}
	respondJSON(w, http.StatusOK, memberships)
}

// GetGroupCsm returns a specific Group resource by label
func GetMembershipCsm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, fmt.Errorf("ID is required"))
		return
	}

	groups, err := plugins.Store.LoadAllGroups(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load groups: %w", err))
		return
	}

	componentEndpoints, err := plugins.Store.LoadAllComponentEndpoints(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to load componentendpoints: %w", err))
		return
	}

	m := componentsToEmptyMemberships(componentEndpoints)
	m = groupsToMemberships(groups, m)

	membership, found := m[id]
	if !found {
		respondError(w, http.StatusNotFound, fmt.Errorf("membership not found: %s", id))
		return
	}

	respondJSON(w, http.StatusOK, membership)
}

func groupsToMemberships(groups []*v1.Group, memberships map[string]*v1.Membership) map[string]*v1.Membership {
	for _, group := range groups {
		for _, memberId := range group.Spec.Members.IDs {
			if v, ok := memberships[memberId]; ok {
				v.GroupLabels = append(v.GroupLabels, group.Spec.Label)
			} else {
				v := &v1.Membership{ID: memberId}
				v.GroupLabels = append(v.GroupLabels, group.Spec.Label)
				memberships[memberId] = v
			}
		}
	}
	return memberships
}

func componentsToEmptyMemberships(components []*v1.ComponentEndpoint) map[string]*v1.Membership {
	m := make(map[string]*v1.Membership)
	for _, c := range components {
		v := &v1.Membership{
			ID:          c.Spec.ID,
			GroupLabels: make([]string, 0)}
		m[c.Spec.ID] = v
	}
	return m
}
