// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package perses

import (
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	"github.com/perses/perses/pkg/model/api/config"
	//corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/util/intstr"
)

func (p *perses) perses() *persesv1alpha1.Perses {
	return &persesv1alpha1.Perses{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Perses",
			APIVersion: "perses.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.values.Name,
			Namespace: p.namespace,
		},
		Spec: persesv1alpha1.PersesSpec{
			Metadata: &persesv1alpha1.Metadata{
				Labels: p.getLabels(),
			},
			Config: persesv1alpha1.PersesConfig{
				Config: config.Config{
					Database: config.Database{
						File: &config.File{
							Folder:    "/perses",
							Extension: "yaml",
						},
					},
					EphemeralDashboard: config.EphemeralDashboard{
						Enable: false,
					},
				},
			},
			ContainerPort: 8080,
			Image:         p.values.Image,
		},
	}
}
