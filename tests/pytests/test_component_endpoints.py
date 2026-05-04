# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json

from conftest import inventory_base_url, print_response


def test_component_endpoints(discover_hardware):
    # Step 1: GET existing component endpoints to use as a template
    response = requests.get(f"{inventory_base_url}/v2/Inventory/ComponentEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET component endpoints: {response.url}")

    component_endpoints = json.loads(response.text).get("ComponentEndpoints", [])
    if not component_endpoints:
        pytest.fail("No existing component endpoints found to use as a template")

    # Step 2: Copy the first spec and change the ID to create a new unique resource
    new_spec = component_endpoints[0].copy()
    new_id = "x9999c0s0b0n0"
    new_spec["ID"] = new_id

    # Step 3: POST the new component endpoint
    post_body = {"ComponentEndpoints": [new_spec]}
    response = requests.post(f"{inventory_base_url}/v2/Inventory/ComponentEndpoints", json=post_body)
    if response.status_code != 201:
        print_response("POST", response)
        pytest.fail(f"Failed to POST component endpoint: expected 201, got {response.status_code}")

    # Step 4: GET the newly created component endpoint and verify it exists
    response = requests.get(f"{inventory_base_url}/v2/Inventory/ComponentEndpoints/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET new component endpoint {new_id}: {response.status_code}")

    created = json.loads(response.text)
    assert created.get("ID") == new_id, \
        f"Expected ID {new_id!r}, got {created.get('ID')!r}"

    # Step 5: Modify the spec
    updated_spec = created.copy()
    updated_spec["Domain"] = "test-domain-9999.local"

    # Step 6: PUT the modified component endpoint
    response = requests.put(f"{inventory_base_url}/v2/Inventory/ComponentEndpoints/{new_id}", json=updated_spec)
    if not response.ok:
        print_response("PUT", response)
        pytest.fail(f"Failed to PUT component endpoint {new_id}: {response.status_code}")

    # Step 7: GET again to verify the update was applied
    response = requests.get(f"{inventory_base_url}/v2/Inventory/ComponentEndpoints/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET updated component endpoint {new_id}: {response.status_code}")

    after_put = json.loads(response.text)
    assert after_put.get("Domain") == "test-domain-9999.local", \
        f"Expected Domain 'test-domain-9999.local' after PUT, got {after_put.get('Domain')!r}"

    # Step 8: DELETE the component endpoint
    response = requests.delete(f"{inventory_base_url}/v2/Inventory/ComponentEndpoints/{new_id}")
    if not response.ok:
        print_response("DELETE", response)
        pytest.fail(f"Failed to DELETE component endpoint {new_id}: {response.status_code}")

    # Step 9: GET again to verify the component endpoint is gone
    response = requests.get(f"{inventory_base_url}/v2/Inventory/ComponentEndpoints/{new_id}")
    assert response.status_code == 404, \
        f"Expected 404 after DELETE, got {response.status_code}"
