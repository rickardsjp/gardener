// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "github.com/gardener/gardener/pkg/client/seedmanagement/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Gardenlets returns a GardenletInformer.
	Gardenlets() GardenletInformer
	// ManagedSeeds returns a ManagedSeedInformer.
	ManagedSeeds() ManagedSeedInformer
	// ManagedSeedSets returns a ManagedSeedSetInformer.
	ManagedSeedSets() ManagedSeedSetInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Gardenlets returns a GardenletInformer.
func (v *version) Gardenlets() GardenletInformer {
	return &gardenletInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ManagedSeeds returns a ManagedSeedInformer.
func (v *version) ManagedSeeds() ManagedSeedInformer {
	return &managedSeedInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ManagedSeedSets returns a ManagedSeedSetInformer.
func (v *version) ManagedSeedSets() ManagedSeedSetInformer {
	return &managedSeedSetInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
