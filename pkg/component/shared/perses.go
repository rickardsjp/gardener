// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"github.com/gardener/gardener/pkg/component"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener/imagevector"
	"github.com/gardener/gardener/pkg/component/observability/monitoring/perses"
)

// NewPerses creates a new prometheus deployer.
func NewPerses(c client.Client, namespace string, values perses.Values) (component.DeployWaiter, error) {
	imagePerses, err := imagevector.Containers().FindImage(imagevector.ContainerImageNamePerses)
	if err != nil {
		return nil, err
	}

	values.Image = imagePerses.String()

	return perses.New(c, namespace, values), nil
}
