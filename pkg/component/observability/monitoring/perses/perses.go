// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package perses

import (
	"fmt"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/utils"
	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
	"github.com/perses/perses/pkg/model/api/config"
	v1 "github.com/perses/perses/pkg/model/api/v1"
	"github.com/perses/perses/pkg/model/api/v1/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (p *perses) perses() *persesv1alpha1.Perses {
	labels := map[string]string{
		v1beta1constants.LabelNetworkPolicyToDNS: v1beta1constants.LabelNetworkPolicyAllowed,
	}

	if p.values.IsSeedCluster {
		labels = utils.MergeStringMaps(labels, p.getLabels(), map[string]string{
			"networking.resources.gardener.cloud/to-prometheus-aggregate-tcp-9090": v1beta1constants.LabelNetworkPolicyAllowed,
			"networking.resources.gardener.cloud/to-prometheus-seed-tcp-9090":      v1beta1constants.LabelNetworkPolicyAllowed,
			"networking.resources.gardener.cloud/to-prometheus-cache-tcp-9090":     v1beta1constants.LabelNetworkPolicyAllowed,
		})
	}

	if p.values.IsGardenCluster {
		labels = utils.MergeStringMaps(labels, p.getLabels(), map[string]string{
			"networking.resources.gardener.cloud/to-prometheus-garden-tcp-9090":   v1beta1constants.LabelNetworkPolicyAllowed,
			"networking.resources.gardener.cloud/to-prometheus-longterm-tcp-9090": v1beta1constants.LabelNetworkPolicyAllowed,
		})
	}

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
				Labels: labels,
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

func (p *perses) persesDatasources() []client.Object {

	datasources := []client.Object{}

	if p.values.IsGardenCluster {
		// add prometheus garden and longterm datasources
		gardenProm := &persesv1alpha1.PersesDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-garden-datasource",
				Namespace: v1beta1constants.GardenNamespace,
			},
			Spec: persesv1alpha1.DatasourceSpec{
				Config: persesv1alpha1.Datasource{
					DatasourceSpec: v1.DatasourceSpec{
						Display: &common.Display{
							Name:        "prometheus-garden",
							Description: "The Garden Prometheus instance",
						},
						Default: false,
						Plugin: common.Plugin{
							Kind: prometheusDatasource,
							// Unfortunately, the CRD "ends" here and goes into a schemaless interface
							Spec: map[string]interface{}{
								"proxy": map[string]interface{}{
									"kind": "HTTPProxy",
									"spec": map[string]interface{}{
										"allowedEndpoints": []interface{}{
											map[string]interface{}{
												"endpointPattern": "/api/v1/labels",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/series",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/metadata",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query_range",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/label/([a-zA-Z0-9_-]+)/values",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/parse_query",
												"method":          "POST",
											},
										},
										"url": fmt.Sprintf("http://%s.%s.svc.cluster.local", "prometheus-garden", v1beta1constants.GardenNamespace),
									},
								},
							},
						},
					},
				},
			},
		}
		longtermProm := &persesv1alpha1.PersesDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-longterm-datasource",
				Namespace: v1beta1constants.GardenNamespace,
			},
			Spec: persesv1alpha1.DatasourceSpec{
				Config: persesv1alpha1.Datasource{
					DatasourceSpec: v1.DatasourceSpec{
						Display: &common.Display{
							Name:        "prometheus-longterm",
							Description: "The long-term Prometheus instance",
						},
						Default: false,
						Plugin: common.Plugin{
							Kind: prometheusDatasource,
							// Unfortunately, the CRD "ends" here and goes into a schemaless interface
							Spec: map[string]interface{}{
								"proxy": map[string]interface{}{
									"kind": "HTTPProxy",
									"spec": map[string]interface{}{
										"allowedEndpoints": []interface{}{
											map[string]interface{}{
												"endpointPattern": "/api/v1/labels",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/series",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/metadata",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query_range",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/label/([a-zA-Z0-9_-]+)/values",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/parse_query",
												"method":          "POST",
											},
										},
										"url": fmt.Sprintf("http://%s.%s.svc.cluster.local", "prometheus-longterm", v1beta1constants.GardenNamespace),
									},
								},
							},
						},
					},
				},
			},
		}
		datasources = append(datasources, gardenProm, longtermProm)
	}

	if p.values.IsSeedCluster {
		// add prometheus cache, aggregate, seed datasources
		seedProm := &persesv1alpha1.PersesDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-seed-datasource",
				Namespace: v1beta1constants.GardenNamespace,
			},
			Spec: persesv1alpha1.DatasourceSpec{
				Config: persesv1alpha1.Datasource{
					DatasourceSpec: v1.DatasourceSpec{
						Display: &common.Display{
							Name:        "prometheus-seed",
							Description: "The Seed Prometheus instance",
						},
						Default: false,
						Plugin: common.Plugin{
							Kind: prometheusDatasource,
							// Unfortunately, the CRD "ends" here and goes into a schemaless interface
							Spec: map[string]interface{}{
								"proxy": map[string]interface{}{
									"kind": "HTTPProxy",
									"spec": map[string]interface{}{
										"allowedEndpoints": []interface{}{
											map[string]interface{}{
												"endpointPattern": "/api/v1/labels",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/series",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/metadata",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query_range",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/label/([a-zA-Z0-9_-]+)/values",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/parse_query",
												"method":          "POST",
											},
										},
										"url": fmt.Sprintf("http://%s.%s.svc.cluster.local", "prometheus-seed", v1beta1constants.GardenNamespace),
									},
								},
							},
						},
					},
				},
			},
		}
		aggregateProm := &persesv1alpha1.PersesDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-aggregate-datasource",
				Namespace: v1beta1constants.GardenNamespace,
			},
			Spec: persesv1alpha1.DatasourceSpec{
				Config: persesv1alpha1.Datasource{
					DatasourceSpec: v1.DatasourceSpec{
						Display: &common.Display{
							Name:        "prometheus-aggregate",
							Description: "The Aggregate Prometheus instance",
						},
						Default: true,
						Plugin: common.Plugin{
							Kind: prometheusDatasource,
							// Unfortunately, the CRD "ends" here and goes into a schemaless interface
							Spec: map[string]interface{}{
								"proxy": map[string]interface{}{
									"kind": "HTTPProxy",
									"spec": map[string]interface{}{
										"allowedEndpoints": []interface{}{
											map[string]interface{}{
												"endpointPattern": "/api/v1/labels",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/series",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/metadata",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query_range",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/label/([a-zA-Z0-9_-]+)/values",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/parse_query",
												"method":          "POST",
											},
										},
										"url": fmt.Sprintf("http://%s.%s.svc.cluster.local", "prometheus-aggregate", v1beta1constants.GardenNamespace),
									},
								},
							},
						},
					},
				},
			},
		}
		cacheProm := &persesv1alpha1.PersesDatasource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-cache-datasource",
				Namespace: v1beta1constants.GardenNamespace,
			},
			Spec: persesv1alpha1.DatasourceSpec{
				Config: persesv1alpha1.Datasource{
					DatasourceSpec: v1.DatasourceSpec{
						Display: &common.Display{
							Name:        "prometheus-cache",
							Description: "The Cache Prometheus instance",
						},
						Default: false,
						Plugin: common.Plugin{
							Kind: prometheusDatasource,
							// Unfortunately, the CRD "ends" here and goes into a schemaless interface
							Spec: map[string]interface{}{
								"proxy": map[string]interface{}{
									"kind": "HTTPProxy",
									"spec": map[string]interface{}{
										"allowedEndpoints": []interface{}{
											map[string]interface{}{
												"endpointPattern": "/api/v1/labels",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/series",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/metadata",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/query_range",
												"method":          "POST",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/label/([a-zA-Z0-9_-]+)/values",
												"method":          "GET",
											},
											map[string]interface{}{
												"endpointPattern": "/api/v1/parse_query",
												"method":          "POST",
											},
										},
										"url": fmt.Sprintf("http://%s.%s.svc.cluster.local", "prometheus-cache", v1beta1constants.GardenNamespace),
									},
								},
							},
						},
					},
				},
			},
		}
		datasources = append(datasources, seedProm, aggregateProm, cacheProm)
	}

	return datasources
}
