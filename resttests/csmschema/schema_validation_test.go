/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Tests that verify the CSM JSON-schema constraints (schemas/csm/) are
 * enforced when the server runs with --schema pointing at that directory.
 *
 * The CSM schemas add stricter constraints beyond the defaults:
 *   components_schema.json       – enum values for Type, State, Flag, NetType, Arch, Class
 *   ethernet_interface_schema.json – pattern for MACAddress, required ID in IPAddress items
 *   hardware_schema.json         – enum values for Type, Status, HWInventoryByLocationType
 *   redfish_endpoint_schema.json – enum values for Type
 *
 * Endpoints under test (non-CSM generic routes, schema overridden by CSM files):
 *   POST /components          — schemas/csm/components_schema.json
 *   POST /ethernetinterfaces  — schemas/csm/ethernet_interface_schema.json
 *   POST /hardwares           — schemas/csm/hardware_schema.json
 *   POST /redfishendpoints    — schemas/csm/redfish_endpoint_schema.json
 */

package csmschema

import (
	"net/http"
	"testing"
)

// ─── Component (/components) ──────────────────────────────────────────────────
// Schema: schemas/csm/components_schema.json
// CSM additions vs default: enum constraints on Type, State, Flag, NetType,
// Arch, and Class.

// TestCsmComponent_MissingSpec verifies the baseline "spec required" rule still
// holds when the CSM schema is loaded.
func TestCsmComponent_MissingSpec(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0n0"},
	})
}

// TestCsmComponent_MissingID verifies that omitting spec.ID is rejected.
func TestCsmComponent_MissingID(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0n0"},
		"spec": map[string]interface{}{
			"Type": "Node",
		},
	})
}

// TestCsmComponent_EmptyID verifies that an empty spec.ID is rejected
// (minLength: 1).
func TestCsmComponent_EmptyID(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": ""},
		"spec": map[string]interface{}{
			"ID":   "",
			"Type": "Node",
		},
	})
}

// TestCsmComponent_InvalidType verifies that a Type value not present in the
// HMSType enum is rejected.  The default schema accepts any string; the CSM
// schema restricts it to known HMSType values.
func TestCsmComponent_InvalidType(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0n0"},
		"spec": map[string]interface{}{
			"ID":   "x3000c0s0b0n0",
			"Type": "NotARealHMSType",
		},
	})
}

// TestCsmComponent_InvalidState verifies that a State value outside the allowed
// enum (Unknown, Empty, Populated, Off, On, Standby, Halt, Ready) is rejected.
func TestCsmComponent_InvalidState(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0n0"},
		"spec": map[string]interface{}{
			"ID":    "x3000c0s0b0n0",
			"Type":  "Node",
			"State": "Rebooting",
		},
	})
}

// TestCsmComponent_InvalidFlag verifies that a Flag value outside the allowed
// enum (OK, Warning, Alert, Locked) is rejected.
func TestCsmComponent_InvalidFlag(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0n0"},
		"spec": map[string]interface{}{
			"ID":   "x3000c0s0b0n0",
			"Type": "Node",
			"Flag": "Critical",
		},
	})
}

// TestCsmComponent_InvalidArch verifies that an Arch value outside the allowed
// enum (X86, ARM, Other) is rejected.
func TestCsmComponent_InvalidArch(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0n0"},
		"spec": map[string]interface{}{
			"ID":   "x3000c0s0b0n0",
			"Type": "Node",
			"Arch": "RISC-V",
		},
	})
}

// TestCsmComponent_InvalidClass verifies that a Class value outside the allowed
// enum (River, Mountain, Hill) is rejected.
func TestCsmComponent_InvalidClass(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0n0"},
		"spec": map[string]interface{}{
			"ID":    "x3000c0s0b0n0",
			"Type":  "Node",
			"Class": "Valley",
		},
	})
}

// TestCsmComponent_InvalidNetType verifies that a NetType value outside the
// allowed enum (Sling, Infiniband, Ethernet, OEM, None) is rejected.
func TestCsmComponent_InvalidNetType(t *testing.T) {
	assertSchemaError(t, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0n0"},
		"spec": map[string]interface{}{
			"ID":      "x3000c0s0b0n0",
			"Type":    "Node",
			"NetType": "WiFi",
		},
	})
}

// TestCsmComponent_ValidCreate verifies that a fully valid component body is
// accepted (HTTP 201) when all enum values are within the allowed sets.
func TestCsmComponent_ValidCreate(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/components", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x5000c0s0b0n0"},
		"spec": map[string]interface{}{
			"ID":      "x5000c0s0b0n0",
			"Type":    "Node",
			"State":   "On",
			"Flag":    "OK",
			"Arch":    "X86",
			"Class":   "Mountain",
			"NetType": "Sling",
		},
	})
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// ─── EthernetInterface (/ethernetinterfaces) ──────────────────────────────────
// Schema: schemas/csm/ethernet_interface_schema.json
// CSM additions vs default: MACAddress must match the colon/hyphen-separated
// hex pattern; IPAddresses items require an "IPAddress" field.

// TestCsmEthernetInterface_MissingSpec verifies the baseline "spec required"
// rule still holds when the CSM schema is loaded.
func TestCsmEthernetInterface_MissingSpec(t *testing.T) {
	assertSchemaError(t, "/ethernetinterfaces", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "a1:00:00:00:00:01"},
	})
}

// TestCsmEthernetInterface_MissingID verifies that omitting spec.ID is rejected.
func TestCsmEthernetInterface_MissingID(t *testing.T) {
	assertSchemaError(t, "/ethernetinterfaces", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "a1:00:00:00:00:01"},
		"spec": map[string]interface{}{
			"MACAddress": "a1:00:00:00:00:01",
		},
	})
}

// TestCsmEthernetInterface_InvalidMAC verifies that a MACAddress value that
// does not match the pattern ^([0-9A-Fa-f]{2}[:-]?){5}([0-9A-Fa-f]{2})$ is
// rejected.  The default schema accepts any string; the CSM schema enforces the
// pattern.
func TestCsmEthernetInterface_InvalidMAC(t *testing.T) {
	assertSchemaError(t, "/ethernetinterfaces", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "bad-mac"},
		"spec": map[string]interface{}{
			"ID":         "bad-mac",
			"MACAddress": "not-a-mac-address",
		},
	})
}

// TestCsmEthernetInterface_InvalidMAC_Short verifies that a truncated MAC
// (fewer than 12 hex digits) is rejected by the pattern.
func TestCsmEthernetInterface_InvalidMAC_Short(t *testing.T) {
	assertSchemaError(t, "/ethernetinterfaces", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "short-mac"},
		"spec": map[string]interface{}{
			"ID":         "short-mac",
			"MACAddress": "a1:b2:c3:d4",
		},
	})
}

// TestCsmEthernetInterface_ValidMAC_Colon verifies that a colon-separated MAC
// address passes the pattern constraint and is accepted.
func TestCsmEthernetInterface_ValidMAC_Colon(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/ethernetinterfaces", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "a1:00:00:00:00:02"},
		"spec": map[string]interface{}{
			"ID":         "a1:00:00:00:00:02",
			"MACAddress": "a1:00:00:00:00:02",
		},
	})
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// TestCsmEthernetInterface_ValidMAC_Hyphen verifies that a hyphen-separated MAC
// address also passes the pattern constraint.
func TestCsmEthernetInterface_ValidMAC_Hyphen(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/ethernetinterfaces", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "a1-00-00-00-00-03"},
		"spec": map[string]interface{}{
			"ID":         "a1-00-00-00-00-03",
			"MACAddress": "a1-00-00-00-00-03",
		},
	})
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// TestCsmEthernetInterface_ValidMAC_NoSeparator verifies that a MAC address
// with no separators (12 contiguous hex digits) also passes the pattern.
func TestCsmEthernetInterface_ValidMAC_NoSeparator(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/ethernetinterfaces", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "a100000000ff"},
		"spec": map[string]interface{}{
			"ID":         "a100000000ff",
			"MACAddress": "a100000000ff",
		},
	})
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// ─── Hardware (/hardwares) ────────────────────────────────────────────────────
// Schema: schemas/csm/hardware_schema.json
// CSM additions vs default: enum constraints on Type, Status, and
// HWInventoryByLocationType.

// TestCsmHardware_MissingSpec verifies the baseline "spec required" rule still
// holds when the CSM schema is loaded.
func TestCsmHardware_MissingSpec(t *testing.T) {
	assertSchemaError(t, "/hardwares", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s5b0n0"},
	})
}

// TestCsmHardware_MissingID verifies that omitting spec.ID is rejected.
func TestCsmHardware_MissingID(t *testing.T) {
	assertSchemaError(t, "/hardwares", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s5b0n0"},
		"spec": map[string]interface{}{
			"Type":   "Node",
			"Status": "Populated",
		},
	})
}

// TestCsmHardware_InvalidType verifies that a Type value not in the HMSType
// enum is rejected.
func TestCsmHardware_InvalidType(t *testing.T) {
	assertSchemaError(t, "/hardwares", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s5b0n0"},
		"spec": map[string]interface{}{
			"ID":   "x3000c0s5b0n0",
			"Type": "SuperNode",
		},
	})
}

// TestCsmHardware_InvalidStatus verifies that a Status value other than
// "Populated" or "Empty" is rejected.
func TestCsmHardware_InvalidStatus(t *testing.T) {
	assertSchemaError(t, "/hardwares", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s5b0n0"},
		"spec": map[string]interface{}{
			"ID":     "x3000c0s5b0n0",
			"Type":   "Node",
			"Status": "Unknown",
		},
	})
}

// TestCsmHardware_InvalidHWInventoryByLocationType verifies that an
// HWInventoryByLocationType value not present in its enum is rejected.
func TestCsmHardware_InvalidHWInventoryByLocationType(t *testing.T) {
	assertSchemaError(t, "/hardwares", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s5b0n0"},
		"spec": map[string]interface{}{
			"ID":                        "x3000c0s5b0n0",
			"Type":                      "Node",
			"HWInventoryByLocationType": "HWInvByLocSuperNode",
		},
	})
}

// TestCsmHardware_ValidCreate verifies that a hardware body with valid enum
// values is accepted (HTTP 201).
func TestCsmHardware_ValidCreate(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/hardwares", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x5000c0s5b0n0"},
		"spec": map[string]interface{}{
			"ID":                        "x5000c0s5b0n0",
			"Type":                      "Node",
			"Status":                    "Populated",
			"HWInventoryByLocationType": "HWInvByLocNode",
		},
	})
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// TestCsmHardware_ValidStatus_Empty verifies that Status "Empty" is accepted.
func TestCsmHardware_ValidStatus_Empty(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/hardwares", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x5000c0s5b0n1"},
		"spec": map[string]interface{}{
			"ID":     "x5000c0s5b0n1",
			"Type":   "Node",
			"Status": "Empty",
		},
	})
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// ─── RedfishEndpoint (/redfishendpoints) ──────────────────────────────────────
// Schema: schemas/csm/redfish_endpoint_schema.json
// CSM additions vs default: enum constraint on Type.

// TestCsmRedfishEndpoint_MissingSpec verifies the baseline "spec required" rule
// still holds when the CSM schema is loaded.
func TestCsmRedfishEndpoint_MissingSpec(t *testing.T) {
	assertSchemaError(t, "/redfishendpoints", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0"},
	})
}

// TestCsmRedfishEndpoint_MissingID verifies that omitting spec.ID is rejected.
func TestCsmRedfishEndpoint_MissingID(t *testing.T) {
	assertSchemaError(t, "/redfishendpoints", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0"},
		"spec": map[string]interface{}{
			"Type":     "NodeBMC",
			"Hostname": "bmc.example.com",
		},
	})
}

// TestCsmRedfishEndpoint_InvalidType verifies that a Type value not in the
// HMSType enum is rejected by the CSM schema.
func TestCsmRedfishEndpoint_InvalidType(t *testing.T) {
	assertSchemaError(t, "/redfishendpoints", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x3000c0s0b0"},
		"spec": map[string]interface{}{
			"ID":   "x3000c0s0b0",
			"Type": "SuperBMC",
		},
	})
}

// TestCsmRedfishEndpoint_ValidCreate verifies that a redfish endpoint body with
// a valid HMSType and all required fields is accepted (HTTP 201).
func TestCsmRedfishEndpoint_ValidCreate(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/redfishendpoints", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "x5000c0s0b0"},
		"spec": map[string]interface{}{
			"ID":       "x5000c0s0b0",
			"Type":     "NodeBMC",
			"Hostname": "bmc5000.example.com",
		},
	})
	requireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}
