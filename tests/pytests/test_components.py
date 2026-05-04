# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json

from conftest import inventory_base_url, print_response


def test_components(discover_hardware):
    # Step 1: GET an existing component to use as a template
    response = requests.get(f"{inventory_base_url}/v2/State/Components")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET components: {response.url}")

    components = json.loads(response.text).get("Components", [])
    if not components:
        pytest.fail("No existing components found to use as a template")

    # Step 2: Copy the first component and modify the ID to create a new resource
    new_spec = components[0].copy()
    new_id = "x9999c0s0b0n0"
    new_spec["ID"] = new_id

    # Step 3: POST the new component
    post_body = {"Components": [new_spec]}
    response = requests.post(f"{inventory_base_url}/v2/State/Components", json=post_body)
    if response.status_code != 201:
        print_response("POST", response)
        pytest.fail(f"Failed to POST component: expected 201, got {response.status_code}")

    # Step 4: GET the newly created component and verify it exists
    response = requests.get(f"{inventory_base_url}/v2/State/Components/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET new component {new_id}: {response.status_code}")

    created = json.loads(response.text)
    assert created.get("spec", {}).get("ID") == new_id, \
        f"Expected ID {new_id!r}, got {created.get('spec', {}).get('ID')!r}"

    # Step 5: Modify the component spec
    updated_spec = created.get("spec", {}).copy()
    updated_spec["State"] = "Off"

    # Step 6: PUT the modified component
    response = requests.put(f"{inventory_base_url}/v2/State/Components/{new_id}", json=updated_spec)
    if not response.ok:
        print_response("PUT", response)
        pytest.fail(f"Failed to PUT component {new_id}: {response.status_code}")

    # Step 7: GET again to verify the update was applied
    response = requests.get(f"{inventory_base_url}/v2/State/Components/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET updated component {new_id}: {response.status_code}")

    after_put = json.loads(response.text)
    assert after_put.get("spec", {}).get("State") == "Off", \
        f"Expected State 'Off' after PUT, got {after_put.get('spec', {}).get('State')!r}"

    # Step 8: DELETE the component
    response = requests.delete(f"{inventory_base_url}/v2/State/Components/{new_id}")
    if not response.ok:
        print_response("DELETE", response)
        pytest.fail(f"Failed to DELETE component {new_id}: {response.status_code}")

    # Step 9: GET again to verify the component is gone
    response = requests.get(f"{inventory_base_url}/v2/State/Components/{new_id}")
    assert response.status_code == 404, \
        f"Expected 404 after DELETE, got {response.status_code}"
