// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

var (
	ProviderSpec = []byte(`
{
	"machineClassRef": {
		"name": "foo"
	},
	"machinePoolSelector": {
			"foo": "bar"
	},
	"machinePoolRef": {
		"name": "foo"
	},
	"networkInterfaces": [
		{
			"ephemeral": {
				"networkInterfaceTemplate": {
					"metadata": {
						"creationTimestamp": null
					},
					"spec": {
						"ipFamilies": [
							"IPv4"
						],
						"ips": [
							{
								"value": "10.0.0.8"
							}
						],
						"networkRef": {
							"name": "network-ref1"
						},
						"virtualIP": {
							"ephemeral": {
								"virtualIPTemplate": {
									"spec": {
										"ipFamily": "IPv4",
										"type": "Public"
									}
								}
							}
						}
					}
				}
			},
			"name": "net-interface"
		}
	],
	"volumes": [
            {
                "device": "oda",
                "name": "root-disk-1",
                "volumeRef": {
                    "name": "machine-0"
                }
            }
        ],
	"ignitionSecret": {
		"name": "foo"
	},
	"image": "foo",
	"imagePullSecretRef": {
		"name": "foo"
	},
	"ignitionSecretKey": "ignition.json",
    "ignition": "passwd:\n  users:\n    - groups: [group1]\n      name: xyz\n      sshAuthorizedKeys: ssh-ed25519 AAABC3NzaC1lZDI1NTE5AAAAIGqrmrq1XwWnPJoSsAeuVcDQNqA5XQK\n      shell: \/bin\/bash"
}
	`)
)
