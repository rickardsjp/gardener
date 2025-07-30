// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package perses

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/component"
	"github.com/gardener/gardener/pkg/utils/managedresources"
)

const (
	ServiceName = "perses"
	Port        = 8080
)

// Values contains configuration values for the Perses resources
type Values struct {
	// Name is the name of the perses instance and the ManagedResource
	Name string
	// Image is the name of the perses image
	Image string
}

// New creates a new instance of DeployWaiter for the perses instance.
func New(client client.Client, namespace string, values Values) component.DeployWaiter {
	return &perses{
		client:    client,
		namespace: namespace,
		values:    values,
	}
}

type perses struct {
	client    client.Client
	namespace string
	values    Values
}

func (p *perses) Deploy(ctx context.Context) error {
	registry := managedresources.NewRegistry(kubernetes.SeedScheme, kubernetes.SeedCodec, kubernetes.SeedSerializer)

	resources, err := registry.AddAllAndSerialize(
		p.perses(),
	)
	if err != nil {
		return err
	}
	return managedresources.CreateForSeed(ctx, p.client, p.namespace, p.name(), false, resources)
}

func (p *perses) Destroy(ctx context.Context) error {
	return managedresources.DeleteForSeed(ctx, p.client, p.namespace, p.name())
}

var TimeoutWaitForManagedResource = 5 * time.Minute

func (p *perses) Wait(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, TimeoutWaitForManagedResource)
	defer cancel()

	return managedresources.WaitUntilHealthy(timeoutCtx, p.client, p.namespace, p.name())
}

func (p *perses) WaitCleanup(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, TimeoutWaitForManagedResource)
	defer cancel()

	return managedresources.WaitUntilDeleted(timeoutCtx, p.client, p.namespace, p.name())
}

func (p *perses) name() string {
	return "perses-" + p.values.Name
}

func (p *perses) getLabels() map[string]string {
	return map[string]string{
		"instance": p.name(),
	}
}
