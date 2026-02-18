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
}
