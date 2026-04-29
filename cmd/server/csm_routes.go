// Copyright © 2025-2026 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT
package main

import "github.com/go-chi/chi/v5"

func RegisterUnprotectedCsmRoutes(r chi.Router) {
	r.Route("/hsm/v2/service/ready", func(r chi.Router) {
		r.Get("/", GetReadinessCsm)
	})
	r.Route("/hsm/v2/service/liveness", func(r chi.Router) {
		r.Get("/", GetLivenessCsm)
	})
	// other smd routes
	// Get /hsm/v2/service/values
	// Get /hsm/v2/service/values/arch
	// Get /hsm/v2/service/values/class
	// Get /hsm/v2/service/values/flag
	// Get /hsm/v2/service/values/nettype
	// Get /hsm/v2/service/values/role
	// Get /hsm/v2/service/values/subrole
	// Get /hsm/v2/service/values/state
	// Get /hsm/v2/service/values/type
}

func RegisterProtectedCsmRoutes(r chi.Router) {

	// Component routes
	r.Route("/hsm/v2/State/Components", func(r chi.Router) {
		r.Get("/", GetComponentsCsm)
		r.Post("/", CreateComponentCsm)
		// r.Delete("/", DeleteAllComponentCsm) // todo (smd has it but maybe not needed)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetComponentCsm)
			r.Put("/", UpdateComponentCsm)
			r.Delete("/", DeleteComponentCsm)
		})
		// other smd routes
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
		r.Get("/", GetComponentEndpointsCsm)
		r.Post("/", CreateComponentEndpointCsm)
		// todo (optional)
		// DELETE /Inventory/ComponentEndpoints
		// r.Delete("/", DeleteAllComponentCsm) // todo (smd has it but it is probably not needed)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetComponentEndpointCsm)
			r.Put("/", UpdateComponentEndpointCsm)
			r.Delete("/", DeleteComponentEndpointCsm)
		})
	})
	// Hardware inventory routes
	r.Route("/hsm/v2/Inventory/Hardware", func(r chi.Router) {
		r.Get("/", GetHardwaresCsm)
		r.Post("/", CreateHardwareCsm)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetHardwareCsm)
			r.Put("/", UpdateHardwareCsm)
			r.Delete("/", DeleteHardwareCsm)
		})
	})
	// ServiceEndpoint routes
	r.Route("/hsm/v2/Inventory/ServiceEndpoints", func(r chi.Router) {
		r.Get("/", GetServiceEndpointsCsm)
		r.Post("/", CreateServiceEndpointCsm)
		// todo (optional)
		// DELETE /Inventory/ServiceEndpoints
		r.Route("/{redfishType}", func(r chi.Router) {
			r.Get("/", GetServiceEndpointCsm)
			r.Route("/RedfishEndpoints/{redfishID}", func(r chi.Router) {
				r.Get("/", GetServiceEndpointByTypeAndIdCsm)
				r.Put("/", UpdateServiceEndpointCsm)
				r.Delete("/", DeleteServiceEndpointCsm)
			})
		})
	})
	// RedfishEndpoint routes
	r.Route("/hsm/v2/Inventory/RedfishEndpoints", func(r chi.Router) {
		r.Get("/", GetRedfishEndpointsCsm)
		r.Post("/", CreateRedfishEndpointCsm)
		// todo (optional)
		// DELETE /Inventory/RedfishEndpoints
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetRedfishEndpointCsm)
			r.Put("/", UpdateRedfishEndpointV2)
			r.Delete("/", DeleteRedfishEndpointV2)
		})
	})
	// EthernetInterface routes
	r.Route("/hsm/v2/Inventory/EthernetInterfaces", func(r chi.Router) {
		r.Get("/", GetEthernetInterfacesCsm)
		r.Post("/", CreateEthernetInterfaceCsm)
		// todo (optional)
		// DELETE /Inventory/EthernetInterfaces
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetEthernetInterfaceCsm)
			r.Put("/", UpdateEthernetInterfaceCsm)
			r.Delete("/", DeleteEthernetInterfaceCsm)
		})
	})
	// Group routes
	r.Route("/hsm/v2/groups", func(r chi.Router) {
		r.Get("/", GetGroupsCsm)
		r.Post("/", CreateGroupCsm)
		r.Route("/{group_label}", func(r chi.Router) {
			r.Get("/", GetGroupCsm)
			r.Put("/", UpdateGroupCsm)
			r.Delete("/", DeleteGroupCsm)
			// GET /groups/{group_label}/members
			// POST /groups/{group_label}/members
			// PUT /groups/{group_label}/members
			// DELETE /groups/{group_label}/members/{member_id}
		})
	})
	// Membership routes
	r.Route("/hsm/v2/memberships", func(r chi.Router) {
		r.Get("/", GetMembershipsCsm)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetMembershipCsm)
		})
	})
}
