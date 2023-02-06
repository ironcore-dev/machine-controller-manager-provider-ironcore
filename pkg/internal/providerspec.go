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
	"labels": {
		"shoot-name": "my-shoot",
		"shoot-namespace": "my-shoot-namespace"
	},
	"machineClassName": "foo",
	"machinePoolName": "foo",
	"networkName": "my-network",
	"prefixName": "my-prefix",
	"rootDisk": {
		"volumeClassName": "foo",
		"size": "10Gi"
	},
	"ignitionSecret": {
		"name": "foo"
	},
	"image": "my-image",
	"ignitionSecretKey": "ignition.json",
    "ignition": "passwd:\n  users:\n    - groups: [group1]\n      name: xyz\n      sshAuthorizedKeys: ssh-ed25519 AAABC3NzaC1lZDI1NTE5AAAAIGqrmrq1XwWnPJoSsAeuVcDQNqA5XQK\n      shell: \/bin\/bash"
}`)

	ProviderSpecWithPoolRef = []byte(`
{
	"labels": {
		"shoot-name": "my-shoot",
		"shoot-namespace": "my-shoot-namespace"
	},
	"machineClassName": "foo",
	"machinePoolName": "foo",
	"networkName": "my-network",
	"prefixName": "my-prefix",
	"rootDisk": {
		"volumeClassName": "foo",
		"volumePoolName": "foo",
		"size": "10Gi"
	},
	"ignitionSecret": {
		"name": "foo"
	},
	"image": "my-image",
	"ignitionSecretKey": "ignition.json",
    "ignition": "passwd:\n  users:\n    - groups: [group1]\n      name: xyz\n      sshAuthorizedKeys: ssh-ed25519 AAABC3NzaC1lZDI1NTE5AAAAIGqrmrq1XwWnPJoSsAeuVcDQNqA5XQK\n      shell: \/bin\/bash"
}`)
)
