// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ironcore

import (
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GetVolumeIDs", func() {
	_, _, drv := SetupTest()

	It("should get volume IDs", func(ctx SpecContext) {
		By("giving correct driver name")
		csiDriverName := IroncoreCSIDriver
		volumeIDs := []string{"vol-ironcore-csi"}
		ret, err := (*drv).GetVolumeIDs(ctx, &driver.GetVolumeIDsRequest{
			PVSpecs: []*corev1.PersistentVolumeSpec{
				{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver:       csiDriverName,
							VolumeHandle: volumeIDs[0],
						},
					},
				},
			},
		})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ret).NotTo(BeNil())
		Expect(ret.VolumeIDs).To(Equal(volumeIDs))
		By("giving wrong driver name")
		csiDriverName = "wrong-driver-name"
		ret, err = (*drv).GetVolumeIDs(ctx, &driver.GetVolumeIDsRequest{
			PVSpecs: []*corev1.PersistentVolumeSpec{
				{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver:       csiDriverName,
							VolumeHandle: volumeIDs[0],
						},
					},
				},
			},
		})
		var emptyVolumeIDs []string
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ret).NotTo(BeNil())
		Expect(ret.VolumeIDs).To(Equal(emptyVolumeIDs))
	})
})
