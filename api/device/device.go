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

package device

import (
	"fmt"
	"regexp"
)

const (
	letters    = "abcdefghijklmnopqrstuvwxyz"
	numLetters = len(letters)

	// MaxIndex is the maximum index usable for Name / returned by ParseName.
	MaxIndex = numLetters*numLetters + numLetters - 1

	// VirtioPrefix is the device prefix used by virtio devices.
	VirtioPrefix = "vd"
)

var (
	nameRegex = regexp.MustCompile("^(?P<prefix>[a-z]{2})(?P<index>[a-z][a-z]?)$")
)

// ParseName parses the name into its prefix and index. An error is returned if the name is not a valid device name.
func ParseName(name string) (string, int, error) {
	match := nameRegex.FindStringSubmatch(name)
	if match == nil {
		return "", 0, fmt.Errorf("%s does not match device name regex %s", name, nameRegex)
	}

	prefix := match[1]

	idxStr := match[2]
	if len(idxStr) == 1 {
		idx := int(idxStr[0] - 'a')
		return prefix, idx, nil
	}

	r1, r2 := int(idxStr[0]-'a'), int(idxStr[1]-'a')
	idx := (r1+1)*numLetters + r2

	return prefix, idx, nil
}
