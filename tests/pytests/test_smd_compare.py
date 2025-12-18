# Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import pytest
import requests

@pytest.fixture()
def discover_hardware():
    # setup
    bmc_nodes = [ "x0c0s1b0", "x0c0s2b0", "x0c0s3b0", "x0c0s4b0" ]

    smd_url = "http://smd:27779/hsm/v2/Inventory/RedfishEndpoints"
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
        response = requests.post(smd_url, json=request_body)
        print(f"Status code: {response.status_code}")
        print("Response: ")
        print(response.json())

    yield

    # tear down


def test_compare(discover_hardware):
    print("compare")

