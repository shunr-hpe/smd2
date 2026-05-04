# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json

from conftest import inventory_base_url, print_response


def test_service_endpoints(discover_hardware):
    # Step 1: GET existing service endpoints to use as a template
    # GET /hsm/v2/Inventory/ServiceEndpoints returns {"ServiceEndpoints": [specs]}
    response = requests.get(f"{inventory_base_url}/v2/Inventory/ServiceEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET service endpoints: {response.url}")

    service_endpoints = json.loads(response.text).get("ServiceEndpoints", [])
    if not service_endpoints:
        pytest.fail("No existing service endpoints found to use as a template")

    # Step 2: Copy the first spec and set unique identifiers.
    # Service endpoints are addressed by RedfishType + RedfishEndpointID in the URL.
    new_spec = service_endpoints[0].copy()
    new_redfish_type = new_spec.get("RedfishType", "AccountService")
    new_redfish_id = "x9999c0s0b0"
    new_spec["RedfishEndpointID"] = new_redfish_id

    # Step 3: POST the new service endpoint
    # POST expects {"ServiceEndpoints": [spec]}
    post_body = {"ServiceEndpoints": [new_spec]}
    response = requests.post(f"{inventory_base_url}/v2/Inventory/ServiceEndpoints", json=post_body)
    if response.status_code != 201:
        print_response("POST", response)
        pytest.fail(f"Failed to POST service endpoint: expected 201, got {response.status_code}")

    # Step 4: GET the newly created service endpoint and verify it exists
    endpoint_url = (
        f"{inventory_base_url}/v2/Inventory/ServiceEndpoints"
        f"/{new_redfish_type}/RedfishEndpoints/{new_redfish_id}"
    )
    response = requests.get(endpoint_url)
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET new service endpoint {new_redfish_type}/{new_redfish_id}: {response.status_code}")

    created = json.loads(response.text)
    assert created.get("RedfishEndpointID") == new_redfish_id, \
        f"Expected RedfishEndpointID {new_redfish_id!r}, got {created.get('RedfishEndpointID')!r}"

    # Step 5: Modify the spec
    updated_spec = created.copy()
    updated_spec["RedfishSubtype"] = "test-subtype-9999"

    # Step 6: PUT the modified service endpoint
    response = requests.put(endpoint_url, json=updated_spec)
    if not response.ok:
        print_response("PUT", response)
        pytest.fail(f"Failed to PUT service endpoint {new_redfish_type}/{new_redfish_id}: {response.status_code}")

    # Step 7: GET again to verify the update was applied
    response = requests.get(endpoint_url)
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to GET updated service endpoint: {response.status_code}")

    after_put = json.loads(response.text)
    assert after_put.get("RedfishSubtype") == "test-subtype-9999", \
        f"Expected RedfishSubtype 'test-subtype-9999' after PUT, got {after_put.get('RedfishSubtype')!r}"

    # Step 8: DELETE the service endpoint
    response = requests.delete(endpoint_url)
    if not response.ok:
        print_response("DELETE", response)
        pytest.fail(f"Failed to DELETE service endpoint {new_redfish_type}/{new_redfish_id}: {response.status_code}")

    # Step 9: GET again to verify the service endpoint is gone
    response = requests.get(endpoint_url)
    assert response.status_code == 404, \
        f"Expected 404 after DELETE, got {response.status_code}"
