# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json

from conftest import inventory_base_url, print_response


def test_hardware(discover_hardware):
    # Step 1: GET existing hardware to use as a template
    # GET /hsm/v2/Inventory/Hardware returns a raw JSON array of hardware specs
    response = requests.get(f"{inventory_base_url}/v2/Inventory/Hardware")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET hardware: {response.url}")

    hardware_list = json.loads(response.text)
    if not hardware_list:
        pytest.fail("No existing hardware found to use as a template")

    # Step 2: Copy the first spec and change the ID to create a new unique resource
    new_spec = hardware_list[0].copy()
    new_id = "x9999c0s0"
    new_spec["ID"] = new_id

    # Step 3: POST the new hardware entry
    # POST expects {"Hardware": [spec]}
    post_body = {"Hardware": [new_spec]}
    response = requests.post(f"{inventory_base_url}/v2/Inventory/Hardware", json=post_body)
    if response.status_code != 201:
        print_response("POST", response)
        pytest.fail(f"Failed to POST hardware: expected 201, got {response.status_code}")

    # Step 4: GET the newly created hardware and verify it exists
    response = requests.get(f"{inventory_base_url}/v2/Inventory/Hardware/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET new hardware {new_id}: {response.status_code}")

    created = json.loads(response.text)
    assert created.get("ID") == new_id, \
        f"Expected ID {new_id!r}, got {created.get('ID')!r}"

    # Step 5: Modify the spec
    updated_spec = created.copy()
    updated_spec["Status"] = "Empty"

    # Step 6: PUT the modified hardware
    response = requests.put(f"{inventory_base_url}/v2/Inventory/Hardware/{new_id}", json=updated_spec)
    if not response.ok:
        print_response("PUT", response)
        pytest.fail(f"Failed to PUT hardware {new_id}: {response.status_code}")

    # Step 7: GET again to verify the update was applied
    response = requests.get(f"{inventory_base_url}/v2/Inventory/Hardware/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET updated hardware {new_id}: {response.status_code}")

    after_put = json.loads(response.text)
    assert after_put.get("Status") == "Empty", \
        f"Expected Status 'Empty' after PUT, got {after_put.get('Status')!r}"

    # Step 8: DELETE the hardware
    response = requests.delete(f"{inventory_base_url}/v2/Inventory/Hardware/{new_id}")
    if not response.ok:
        print_response("DELETE", response)
        pytest.fail(f"Failed to DELETE hardware {new_id}: {response.status_code}")

    # Step 9: GET again to verify the hardware is gone
    response = requests.get(f"{inventory_base_url}/v2/Inventory/Hardware/{new_id}")
    assert response.status_code == 404, \
        f"Expected 404 after DELETE, got {response.status_code}"
