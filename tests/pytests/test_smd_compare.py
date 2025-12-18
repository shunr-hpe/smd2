# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests
import json

smd_base_url = "http://smd:27779/hsm"
smd2_base_url = "http://smd2:8080/hsm"

@pytest.fixture()
def discover_hardware():
    # setup

    headers = {
    }

    bmc_nodes = [ "x0c0s1b0", "x0c0s2b0", "x0c0s3b0", "x0c0s4b0" ]
    for node in bmc_nodes:
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
        print(f"SMD POST RedfishEndpoints: Status code: {response.status_code}")
        print("SMD POST RedfishEndpoints Response: ")
        print(response.json())

    print("POST Components to SMD2")
    response = requests.get(f"{smd_base_url}/v2/State/Components")
    print(f"SMD GET Components: Status code: {response.status_code}")
    print("SMD GET Components Response: ")
    print(response.text)
    smd_components = json.loads(response.text)

    # post components to smd2
    response = requests.post(f"{smd2_base_url}/v2/State/Components", json=smd_components)
    print(f"SMD2 POST Components: Status code: {response.status_code}")
    print("SDM2 POST Components Response: ")
    print(response.text)

    # assert False

    yield

    # tear down


def test_compare(discover_hardware):
    print("compare")

