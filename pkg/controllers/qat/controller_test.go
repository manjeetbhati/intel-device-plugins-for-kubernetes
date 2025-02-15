// Copyright 2021 Intel Corporation. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package qat contains QAT specific reconciliation logic.
package qat

import (
	"reflect"
	"testing"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devicepluginv1 "github.com/intel/intel-device-plugins-for-kubernetes/pkg/apis/deviceplugin/v1"
)

const appLabel = "intel-qat-plugin"

// newDaemonSetExpected creates plugin daemonset
// it's copied from the original controller code (before the usage of go:embed).
func (c *controller) newDaemonSetExpected(rawObj client.Object) *apps.DaemonSet {
	devicePlugin := rawObj.(*devicepluginv1.QatDevicePlugin)
	yes := true
	no := false
	pluginAnnotations := devicePlugin.ObjectMeta.DeepCopy().Annotations

	return &apps.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.ns,
			Name:      "intel-qat-plugin",
			Labels: map[string]string{
				"app": appLabel,
			},
			Annotations: pluginAnnotations,
		},
		Spec: apps.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": appLabel,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": appLabel,
					},
					Annotations: pluginAnnotations,
				},
				Spec: v1.PodSpec{
					AutomountServiceAccountToken: &no,
					Containers: []v1.Container{
						{
							Name:            appLabel,
							Args:            getPodArgs(devicePlugin),
							Image:           devicePlugin.Spec.Image,
							ImagePullPolicy: "IfNotPresent",
							SecurityContext: &v1.SecurityContext{
								SELinuxOptions: &v1.SELinuxOptions{
									Type: "container_device_plugin_t",
								},
								ReadOnlyRootFilesystem:   &yes,
								AllowPrivilegeEscalation: &no,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "devdir",
									MountPath: "/dev/vfio",
									ReadOnly:  true,
								},
								{
									Name:      "debugfsdir",
									MountPath: "/sys/kernel/debug",
									ReadOnly:  true,
								},
								{
									Name:      "pcidir",
									MountPath: "/sys/bus/pci",
								},
								{
									Name:      "kubeletsockets",
									MountPath: "/var/lib/kubelet/device-plugins",
								},
							},
						},
					},
					NodeSelector: map[string]string{"kubernetes.io/arch": "amd64"},
					Volumes: []v1.Volume{
						{
							Name: "devdir",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/dev/vfio",
								},
							},
						},
						{
							Name: "debugfsdir",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/sys/kernel/debug",
								},
							},
						},
						{
							Name: "pcidir",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/sys/bus/pci",
								},
							},
						},
						{
							Name: "kubeletsockets",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/device-plugins",
								},
							},
						},
					},
				},
			},
		},
	}
}

// Test that QAT daemonset created by using go:embed is
// equal to the expected daemonset.
func TestNewDaemonSetQAT(t *testing.T) {
	c := &controller{}

	plugin := &devicepluginv1.QatDevicePlugin{}
	expected := c.newDaemonSetExpected(plugin)
	actual := c.NewDaemonSet(plugin)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected and actuall daemonsets differ: %+s", diff.ObjectGoPrintDiff(expected, actual))
	}
}
