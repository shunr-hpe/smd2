# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json

from conftest import inventory_base_url, print_response


def test_redfish_endpoints(discover_hardware):
    # Step 1: GET existing redfish endpoints to use as a template
    # GET /hsm/v2/Inventory/RedfishEndpoints returns {"RedfishEndpoints": [specs]}
    response = requests.get(f"{inventory_base_url}/v2/Inventory/RedfishEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET redfish endpoints: {response.url}")

    redfish_endpoints = json.loads(response.text).get("RedfishEndpoints", [])
    if not redfish_endpoints:
        pytest.fail("No existing redfish endpoints found to use as a template")

    # Step 2: Copy the first spec and change the ID to create a new unique resource
    new_spec = redfish_endpoints[0].copy()
    new_id = "x9999c0s0b0"
    new_spec["ID"] = new_id
    new_spec["FQDN"] = new_id

    # Step 3: POST the new redfish endpoint
    # POST accepts a single spec (no array wrapper)
    response = requests.post(f"{inventory_base_url}/v2/Inventory/RedfishEndpoints", json=new_spec)
    if response.status_code != 201:
        print_response("POST", response)
        pytest.fail(f"Failed to POST redfish endpoint: expected 201, got {response.status_code}")

    # Step 4: GET the newly created redfish endpoint and verify it exists
    response = requests.get(f"{inventory_base_url}/v2/Inventory/RedfishEndpoints/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET new redfish endpoint {new_id}: {response.status_code}")

    created = json.loads(response.text)
    assert created.get("ID") == new_id, \
        f"Expected ID {new_id!r}, got {created.get('ID')!r}"

    # Step 5: Modify the spec
    updated_spec = created.copy()
    updated_spec["Enabled"] = not updated_spec.get("Enabled", True)

    # Step 6: PUT the modified redfish endpoint
    response = requests.put(f"{inventory_base_url}/v2/Inventory/RedfishEndpoints/{new_id}", json=updated_spec)
    if not response.ok:
        print_response("PUT", response)
        pytest.fail(f"Failed to PUT redfish endpoint {new_id}: {response.status_code}")

    # Step 7: GET again to verify the update was applied
    response = requests.get(f"{inventory_base_url}/v2/Inventory/RedfishEndpoints/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET updated redfish endpoint {new_id}: {response.status_code}")

    after_put = json.loads(response.text)
    assert after_put.get("Enabled") == updated_spec["Enabled"], \
        f"Expected Enabled={updated_spec['Enabled']!r} after PUT, got {after_put.get('Enabled')!r}"

    # Step 8: DELETE the redfish endpoint
    response = requests.delete(f"{inventory_base_url}/v2/Inventory/RedfishEndpoints/{new_id}")
    if not response.ok:
        print_response("DELETE", response)
        pytest.fail(f"Failed to DELETE redfish endpoint {new_id}: {response.status_code}")

    # Step 9: GET again to verify the redfish endpoint is gone
    response = requests.get(f"{inventory_base_url}/v2/Inventory/RedfishEndpoints/{new_id}")
    assert response.status_code == 404, \
        f"Expected 404 after DELETE, got {response.status_code}"
