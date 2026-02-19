# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json
import time
import pprint
from deepdiff import DeepDiff

smd_base_url = "http://smd:27779/hsm"
smd2_base_url = "http://smd2:8080/hsm"

@pytest.fixture()
def discover_hardware():
    # setup

    smd_base_url = "http://smd:27779/hsm"
    smd2_base_url = "http://smd2:8080/hsm"

    headers = {
    }

    bmc_nodes = [ "x0c0s1b0", "x0c0s2b0", "x0c0s3b0", "x0c0s4b0" ]

    response = requests.get(f"{smd_base_url}/v2/Inventory/RedfishEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"Failed to get {response.url}")

    discovered_nodes = [ endpoint.get("ID")
                        for endpoint in json.loads(response.text).get("RedfishEndpoints", [])
                        if endpoint.get("DiscoveryInfo", {}).get("LastDiscoveryStatus") == "DiscoverOK"]
    undiscovered_nodes = list(set(bmc_nodes) - set(discovered_nodes))
    print(f"bmc_nodes: {bmc_nodes}")
    print(f"discovered_nodes: {discovered_nodes}")
    print(f"undiscovered_nodes: {undiscovered_nodes}")


    for node in undiscovered_nodes:
        print(f"discover: {node}")
        request_body = {
                "RedfishEndpoints" : [
                    {
                     "ID" : node,
                     "FQDN" : node,
                     "RediscoverOnUpdate" : True,
                     "User" : "root",
                     "Password" : "root_password"
                     }]
                }
        response = requests.post(f"{smd_base_url}/v2/Inventory/RedfishEndpoints", json=request_body)
        if not response.ok:
            print_response("POST", response)

    if undiscovered_nodes:
        for i in range(0, 10):
            done = True
            print(f"Waiting for discovery to finish. {i}")
            response = requests.get(f"{smd_base_url}/v2/Inventory/RedfishEndpoints")
            if response.ok:
                endpoints = json.loads(response.text)
                discovery_results = { endpoint.get("ID"): endpoint.get("DiscoveryInfo").get("LastDiscoveryStatus")
                                      for endpoint in endpoints.get("RedfishEndpoints")}
                pprint.pprint(discovery_results)
                for node in undiscovered_nodes:
                    endpoint = discovery_results.get(node)
                    print(f"{node} {endpoint}")
                    if endpoint != "DiscoverOK":
                        print(f"- {node} {endpoint}")
                        done = False
                if done:
                    break
            time.sleep(1)

    replicate_components()
    replicate_component_endpoints()
    replicate_ethernet_interfaces()
    replicate_redfish_endpoints()
    replicate_service_endpoints()

    yield

    # tear down


def replicate_components():
    response = requests.get(f"{smd_base_url}/v2/State/Components")
    if not response.ok:
        print_response("GET", response)
    smd_components = json.loads(response.text)

    print("POST Components to SMD2")
    response = requests.post(f"{smd2_base_url}/v2/State/Components", json=smd_components)
    if not response.ok:
        print_response("POST", response)


def replicate_component_endpoints():
    response = requests.get(f"{smd_base_url}/v2/Inventory/ComponentEndpoints")
    if not response.ok:
        print_response("GET", response)
    smd_components = json.loads(response.text)

    print("POST ComponentEndpoints to SMD2")
    response = requests.post(f"{smd2_base_url}/v2/Inventory/ComponentEndpoints", json=smd_components)
    if not response.ok:
        print_response("POST", response)


def replicate_ethernet_interfaces():
    response = requests.get(f"{smd_base_url}/v2/Inventory/EthernetInterfaces")
    if not response.ok:
        print_response("GET", response)
    ethernet_interfaces = json.loads(response.text)

    print("POST EthernetInterfaces to SMD2")
    for eth in ethernet_interfaces:
        response = requests.post(f"{smd2_base_url}/v2/Inventory/EthernetInterfaces", json=eth)
        if not response.ok:
            print_response("POST", response)


def replicate_redfish_endpoints():
    response = requests.get(f"{smd_base_url}/v2/Inventory/RedfishEndpoints")
    if not response.ok:
        print_response("GET", response)
    redfish_endpoints  = json.loads(response.text)

    print("POST RedfishEndpoints to SMD2")
    for redfish_endpoint in redfish_endpoints.get("RedfishEndpoints"):
        response = requests.post(f"{smd2_base_url}/v2/Inventory/RedfishEndpoints", json=redfish_endpoint)
        if not response.ok:
            print_response("POST", response)


def replicate_service_endpoints():
    response = requests.get(f"{smd_base_url}/v2/Inventory/ServiceEndpoints")
    if not response.ok:
        print_response("GET", response)
    smd_service_endpoints = json.loads(response.text)

    print("POST ServiceEndpoints to SMD2")
    response = requests.post(f"{smd2_base_url}/v2/Inventory/ServiceEndpoints", json=smd_service_endpoints)
    if not response.ok:
        print_response("POST", response)


def test_compare_components(discover_hardware):
    # /State/Components
    response = requests.get(f"{smd_base_url}/v2/State/Components")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_components = json.loads(response.text)

    response = requests.get(f"{smd2_base_url}/v2/State/Components")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    smd2_components = json.loads(response.text)

    diff = compare(smd_components.get("Components"), smd2_components.get("Components"))
    if diff:
        pytest.fail(f"The Component list from SMD does not match the list from SMD2. diff: {diff}")

    # /Inventory/ComponentEndpoints
    response = requests.get(f"{smd_base_url}/v2/Inventory/ComponentEndpoints")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_component_endpoints = json.loads(response.text)

    response = requests.get(f"{smd2_base_url}/v2/Inventory/ComponentEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    smd2_component_endpoints = json.loads(response.text)

    diff = compare(smd_component_endpoints.get("ComponentEndpoints"), smd2_component_endpoints.get("ComponentEndpoints"))
    if diff:
        pytest.fail(f"The ComponentEndpoint list from SMD does not match the list from SMD2. diff: {diff}")

    # /Inventory/EthernetInterfaces
    response = requests.get(f"{smd_base_url}/v2/Inventory/EthernetInterfaces")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_ethernet_interfaces = json.loads(response.text)

    response = requests.get(f"{smd2_base_url}/v2/Inventory/EthernetInterfaces")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    smd2_ethernet_interfaces = json.loads(response.text)

    # todo
    # diff = compare(smd_ethernet_interfaces, smd2_ethernet_interfaces, exclude_paths=["root['LastUpdate']"])
    # if diff:
        # pytest.fail(f"The EthernetInterfaces list from SMD does not match the list from SMD2. diff: {diff}")

    # /Inventory/RedfishEndpoints
    response = requests.get(f"{smd_base_url}/v2/Inventory/RedfishEndpoints")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_component_endpoints = json.loads(response.text)

    response = requests.get(f"{smd2_base_url}/v2/Inventory/RedfishEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    smd2_redfish_endpoints = json.loads(response.text)

    diff = compare(smd_component_endpoints.get("RedfishEndpoints"), smd2_redfish_endpoints.get("RedfishEndpoints"))
    if diff:
        pytest.fail(f"The RedfishEndpoint list from SMD does not match the list from SMD2. diff: {diff}")

    # /Inventory/ServiceEndpoints
    response = requests.get(f"{smd_base_url}/v2/Inventory/ServiceEndpoints")
    if response.status_code != 200:
        print_response("GET", response)
        pytest.fail(f" get {response.url}, code: {response.status_code}")

    smd_component_endpoints = json.loads(response.text)

    response = requests.get(f"{smd2_base_url}/v2/Inventory/ServiceEndpoints")
    if not response.ok:
        print_response("GET", response)
        pytest.fail(f"get {response.url}, code: {response.status_code}")

    smd2_redfish_endpoints = json.loads(response.text)

    diff = compare(smd_component_endpoints.get("ServiceEndpoints"), smd2_redfish_endpoints.get("ServiceEndpoints"))
    if diff:
        pytest.fail(f"The ServiceEndpoint list from SMD does not match the list from SMD2. diff: {diff}")


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
