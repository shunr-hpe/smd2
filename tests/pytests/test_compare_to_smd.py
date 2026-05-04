# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json
from deepdiff import DeepDiff

from conftest import smd_base_url, inventory_base_url, print_response


def test_compare_components(discover_hardware):
    response = requests.get(f"{smd_base_url}/v2/State/Components")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_components = json.loads(response.text)

    response = requests.get(f"{inventory_base_url}/v2/State/Components")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    inventory_components = json.loads(response.text)

    diff = compare(smd_components.get("Components"), inventory_components.get("Components"))
    if diff:
        pytest.fail(f"The Component list from SMD does not match the list from the inventory service. diff: {diff}")


def test_compare_component_endpoints(discover_hardware):
    response = requests.get(f"{smd_base_url}/v2/Inventory/ComponentEndpoints")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_component_endpoints = json.loads(response.text)

    response = requests.get(f"{inventory_base_url}/v2/Inventory/ComponentEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    inventory_component_endpoints = json.loads(response.text)

    diff = compare(smd_component_endpoints.get("ComponentEndpoints"), inventory_component_endpoints.get("ComponentEndpoints"))
    if diff:
        pytest.fail(f"The ComponentEndpoint list from SMD does not match the list from the inventory service. diff: {diff}")


def test_compare_ethernet_interfaces(discover_hardware):
    response = requests.get(f"{smd_base_url}/v2/Inventory/EthernetInterfaces")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_ethernet_interfaces = json.loads(response.text)

    response = requests.get(f"{inventory_base_url}/v2/Inventory/EthernetInterfaces")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    inventory_ethernet_interfaces = json.loads(response.text)

    diff = compare(smd_ethernet_interfaces, inventory_ethernet_interfaces, exclude_paths=["root['LastUpdate']"])
    if diff:
        pytest.fail(f"The EthernetInterfaces list from SMD does not match the list from the inventory service. diff: {diff}")


def test_compare_redfish_endpoints(discover_hardware):
    response = requests.get(f"{smd_base_url}/v2/Inventory/RedfishEndpoints")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_redfish_endpoints = json.loads(response.text)

    response = requests.get(f"{inventory_base_url}/v2/Inventory/RedfishEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    inventory_redfish_endpoints = json.loads(response.text)

    diff = compare(smd_redfish_endpoints.get("RedfishEndpoints"), inventory_redfish_endpoints.get("RedfishEndpoints"))
    if diff:
        pytest.fail(f"The RedfishEndpoint list from SMD does not match the list from the inventory service. diff: {diff}")


def test_compare_service_endpoints(discover_hardware):
    response = requests.get(f"{smd_base_url}/v2/Inventory/ServiceEndpoints")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_service_endpoints = json.loads(response.text)

    response = requests.get(f"{inventory_base_url}/v2/Inventory/ServiceEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    inventory_service_endpoints = json.loads(response.text)

    diff = compare(smd_service_endpoints.get("ServiceEndpoints"), inventory_service_endpoints.get("ServiceEndpoints"))
    if diff:
        pytest.fail(f"The ServiceEndpoint list from SMD does not match the list from the inventory service. diff: {diff}")


def test_compare_hardware(discover_hardware):
    response = requests.get(f"{smd_base_url}/v2/Inventory/Hardware")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_hardware = json.loads(response.text)

    response = requests.get(f"{inventory_base_url}/v2/Inventory/Hardware")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    inventory_hardware = json.loads(response.text)

    diff = compare(smd_hardware, inventory_hardware)
    if diff:
        pytest.fail(f"The Hardware list from SMD does not match the list from the inventory service. diff: {diff}")


def get_discovered_nodes(redfishEndpoints):
    discovered_nodes = [ endpoint.get("ID")
                        for endpoint in redfishEndpoints.get("RedfishEndpoints", [])
                        if endpoint.get("DiscoveryInfo", {}).get("LastDiscoveryStatus") == "DiscoveryOK"]


def compare(expected, actual, exclude_paths=None):
    if exclude_paths is None:
        exclude_paths = []

    print("---")
    print(json.dumps(expected, indent=4))
    print("---")
    print(json.dumps(actual, indent=4))

    diff = DeepDiff(expected, actual, group_by="ID", exclude_paths=exclude_paths)
    print("---")
    print(diff)
    print("---")

    return diff


def print_response(method, response):
        print(f"{method} URL: {response.url}, Code: {response.status_code}, Body:")
        print(response.text)
        print(json.dumps(response.text, indent=4))
