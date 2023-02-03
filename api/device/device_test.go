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

package device_test

import (
	"fmt"

	. "github.com/onmetal/machine-controller-manager-provider-onmetal/api/device"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Device", func() {
	DescribeTable("ParseName",
		func(name string, matchers ...types.GomegaMatcher) {
			prefix, idx, err := ParseName(name)
			switch len(matchers) {
			case 1:
				Expect(err).To(matchers[0])
			case 2:
				Expect(err).NotTo(HaveOccurred())
				Expect(prefix).To(matchers[0])
				Expect(idx).To(matchers[1])
			default:
				Fail(fmt.Sprintf("invalid number of matchers: %d, expected 1 (error case) / 2 (success case)", len(matchers)))
			}
		},
		Entry("minimum index", "sda", Equal("sd"), Equal(0)),
		Entry("other prefix", "fda", Equal("fd"), Equal(0)),
		Entry("simple valid name", "sdb", Equal("sd"), Equal(1)),
		Entry("two-letter name", "sdbb", Equal("sd"), Equal(53)),
		Entry("maximum index", "sdzz", Equal("sd"), Equal(MaxIndex)),
		Entry("more than two letters", "sdaaa", HaveOccurred()),
		Entry("invalid prefix", "f_aa", HaveOccurred()),
	)
})
