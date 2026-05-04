# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json

from conftest import inventory_base_url, print_response


def test_groups(discover_hardware):
    # Step 1: Build a new group spec from scratch.
    # Groups are identified by their label, so we use a unique label.
    new_label = "test-group-9999"
    new_spec = {
        "label": new_label,
        "description": "test group created by test_groups",
        "tags": [],
        "members": {"ids": []},
    }

    # Step 2: POST the new group
    # POST accepts a single GroupSpec (no array wrapper)
    response = requests.post(f"{inventory_base_url}/v2/groups", json=new_spec)
    if response.status_code != 201:
        print_response("POST", response)
        pytest.fail(f"Failed to POST group: expected 201, got {response.status_code}")

    # Step 3: GET the newly created group and verify it exists
    response = requests.get(f"{inventory_base_url}/v2/groups/{new_label}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET new group {new_label!r}: {response.status_code}")

    created = json.loads(response.text)
    assert created.get("label") == new_label, \
        f"Expected label {new_label!r}, got {created.get('label')!r}"

    # Step 4: Modify the spec
    updated_spec = created.copy()
    updated_spec["description"] = "updated description 9999"

    # Step 5: PUT the modified group
    response = requests.put(f"{inventory_base_url}/v2/groups/{new_label}", json=updated_spec)
    if not response.ok:
        print_response("PUT", response)
        pytest.fail(f"Failed to PUT group {new_label!r}: {response.status_code}")

    # Step 6: GET again to verify the update was applied
    response = requests.get(f"{inventory_base_url}/v2/groups/{new_label}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET updated group {new_label!r}: {response.status_code}")

    after_put = json.loads(response.text)
    assert after_put.get("description") == "updated description 9999", \
        f"Expected description 'updated description 9999' after PUT, got {after_put.get('description')!r}"

    # Step 7: DELETE the group
    response = requests.delete(f"{inventory_base_url}/v2/groups/{new_label}")
    if not response.ok:
        print_response("DELETE", response)
        pytest.fail(f"Failed to DELETE group {new_label!r}: {response.status_code}")

    # Step 8: GET again to verify the group is gone
    response = requests.get(f"{inventory_base_url}/v2/groups/{new_label}")
    assert response.status_code == 404, \
        f"Expected 404 after DELETE, got {response.status_code}"
