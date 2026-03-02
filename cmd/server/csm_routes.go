// Copyright © 2025-2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
package main

import "github.com/go-chi/chi/v5"

func RegisterSmdV2Routes(r chi.Router) {

	// Component routes
	r.Route("/hsm/v2/State/Components", func(r chi.Router) {
		r.Get("/", GetComponentsSmdV2)
		r.Post("/", CreateComponentSmdV2)
		// r.Delete("/", DeleteAllComponentSmdV2) // todo (smd has it but maybe not needed)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetComponentSmdV2)
			r.Put("/", UpdateComponentSmdV2)
			r.Delete("/", DeleteComponentSmdV2)
		})
		// possible todo
		// Get /State/Components/ByNID/{nid}
		// Get /State/Components/Query/{xname}
		// Post /State/Components/Query
		// Post /State/Components/ByNID/Query
		// Patch /State/Components/BulkStateData
		// Patch /State/Components/{xname}/StateData
		// Patch /State/Components/{xname}/FlagOnly
		// Patch /State/Components/{xname}/Enabled
		// Patch /State/Components/{xname}/SoftwareStatus
		// Patch /State/Components/{xname}/Role
		// Patch /State/Components/{xname}/NID
		// Patch /State/Components/BulkStateData
		// Patch /State/Components/BulkFlagOnly
		// Patch /State/Components/BulkEnabled
		// Patch /State/Components/BulkSoftwareStatus
		// Patch /State/Components/BulkRole
		// Patch /State/Components/BulkNID
	})
	// ComponentEndpoint routes
	r.Route("/hsm/v2/Inventory/ComponentEndpoints", func(r chi.Router) {
		r.Get("/", GetComponentEndpointsSmdV2)
		r.Post("/", CreateComponentEndpointSmdV2)
		// todo (optional)
		// DELETE /Inventory/ComponentEndpoints
		// r.Delete("/", DeleteAllComponentSmdV2) // todo (smd has it but it is probably not needed)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetComponentEndpointSmdV2)
			r.Put("/", UpdateComponentEndpointSmdV2)
			r.Delete("/", DeleteComponentEndpointSmdV2)
		})
	})
	// ServiceEndpoint routes
	r.Route("/hsm/v2/Inventory/ServiceEndpoints", func(r chi.Router) {
		r.Get("/", GetServiceEndpointsSmdV2)
		r.Post("/", CreateServiceEndpointSmdV2)
		// todo (optional)
		// DELETE /Inventory/ServiceEndpoints
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetServiceEndpointSmdV2)
			r.Put("/", UpdateServiceEndpointSmdV2)
			r.Delete("/", DeleteServiceEndpointSmdV2)
			// GET /ServiceEndpoints/{id}/RedfishEndpoints/{redfish_endpoint_id}
			// DELETE /ServiceEndpoints/{id}/RedfishEndpoints/{redfish_endpoint_id}
		})
	})
	// RedfishEndpoint routes
	r.Route("/hsm/v2/Inventory/RedfishEndpoints", func(r chi.Router) {
		r.Get("/", GetRedfishEndpointsSmdV2)
		r.Post("/", CreateRedfishEndpointSmdV2)
		// todo (optional)
		// DELETE /Inventory/RedfishEndpoints
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetRedfishEndpointSmdV2)
			r.Put("/", UpdateRedfishEndpointV2)
			r.Delete("/", DeleteRedfishEndpointV2)
		})
	})
	// EthernetInterface routes
	r.Route("/hsm/v2/Inventory/EthernetInterfaces", func(r chi.Router) {
		r.Get("/", GetEthernetInterfacesSmdV2)
		r.Post("/", CreateEthernetInterfaceSmdV2)
		// todo (optional)
		// DELETE /Inventory/EthernetInterfaces
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetEthernetInterfaceSmdV2)
			r.Put("/", UpdateEthernetInterfaceSmdV2)
			r.Delete("/", DeleteEthernetInterfaceSmdV2)
		})
	})
	// Group routes
	r.Route("/hsm/v2/groups", func(r chi.Router) {
		r.Get("/", GetGroupsSmdV2)
		r.Post("/", CreateGroupSmdV2)
		r.Route("/{group_label}", func(r chi.Router) {
			r.Get("/", GetGroupSmdV2)
			r.Put("/", UpdateGroupSmdV2)
			r.Delete("/", DeleteGroupSmdV2)
			// GET /groups/{group_label}/members
			// POST /groups/{group_label}/members
			// PUT /groups/{group_label}/members
			// DELETE /groups/{group_label}/members/{member_id}
		})
	})
}
