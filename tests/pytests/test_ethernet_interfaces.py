# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json

from conftest import inventory_base_url, print_response


def test_ethernet_interfaces(discover_hardware):
    # Step 1: GET existing ethernet interfaces to use as a template
    # GET /hsm/v2/Inventory/EthernetInterfaces returns a raw JSON array of specs
    response = requests.get(f"{inventory_base_url}/v2/Inventory/EthernetInterfaces")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET ethernet interfaces: {response.url}")

    ethernet_interfaces = json.loads(response.text)
    if not ethernet_interfaces:
        pytest.fail("No existing ethernet interfaces found to use as a template")

    # Step 2: Copy the first spec and change the ID to create a new unique resource
    new_spec = ethernet_interfaces[0].copy()
    new_id = "b42e999abe9f"
    new_spec["ID"] = new_id
    new_spec["MACAddress"] = new_id

    # Step 3: POST the new ethernet interface
    # POST accepts a single spec (no array wrapper)
    response = requests.post(f"{inventory_base_url}/v2/Inventory/EthernetInterfaces", json=new_spec)
    if response.status_code != 201:
        print_response("POST", response)
        pytest.fail(f"Failed to POST ethernet interface: expected 201, got {response.status_code}")

    # Step 4: GET the newly created ethernet interface and verify it exists
    response = requests.get(f"{inventory_base_url}/v2/Inventory/EthernetInterfaces/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET new ethernet interface {new_id}: {response.status_code}")

    created = json.loads(response.text)
    assert created.get("ID") == new_id, \
        f"Expected ID {new_id!r}, got {created.get('ID')!r}"

    # Step 5: Modify the spec
    updated_spec = created.copy()
    updated_spec["Description"] = "test-description-9999"

    # Step 6: PUT the modified ethernet interface
    response = requests.put(f"{inventory_base_url}/v2/Inventory/EthernetInterfaces/{new_id}", json=updated_spec)
    if not response.ok:
        print_response("PUT", response)
        pytest.fail(f"Failed to PUT ethernet interface {new_id}: {response.status_code}")

    # Step 7: GET again to verify the update was applied
    response = requests.get(f"{inventory_base_url}/v2/Inventory/EthernetInterfaces/{new_id}")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET updated ethernet interface {new_id}: {response.status_code}")

    after_put = json.loads(response.text)
    assert after_put.get("Description") == "test-description-9999", \
        f"Expected Description 'test-description-9999' after PUT, got {after_put.get('Description')!r}"

    # Step 8: DELETE the ethernet interface
    response = requests.delete(f"{inventory_base_url}/v2/Inventory/EthernetInterfaces/{new_id}")
    if not response.ok:
        print_response("DELETE", response)
        pytest.fail(f"Failed to DELETE ethernet interface {new_id}: {response.status_code}")

    # Step 9: GET again to verify the ethernet interface is gone
    response = requests.get(f"{inventory_base_url}/v2/Inventory/EthernetInterfaces/{new_id}")
    assert response.status_code == 404, \
        f"Expected 404 after DELETE, got {response.status_code}"
