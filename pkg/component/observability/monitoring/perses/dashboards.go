// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package perses

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	persesv1alpha2 "github.com/perses/perses-operator/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apivalidation "k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var (
	//go:embed dashboards/garden
	gardenDashboards embed.FS
	//go:embed dashboards/seed
	seedDashboards embed.FS
	//go:embed dashboards/garden-seed
	gardenAndSeedDashboards embed.FS
	//go:embed dashboards/garden-shoot
	gardenAndShootDashboards embed.FS
	//go:embed dashboards/common
	commonDashboards embed.FS

	gardenDashboardsPath         = filepath.Join("dashboards", "garden")
	seedDashboardsPath           = filepath.Join("dashboards", "seed")
	gardenAndSeedDashboardsPath  = filepath.Join("dashboards", "garden-seed")
	gardenAndShootDashboardsPath = filepath.Join("dashboards", "garden-shoot")
	commonDashboardsPath         = filepath.Join("dashboards", "common")
	commonVpaDashboardsPath      = filepath.Join(commonDashboardsPath, "vpa")
)

func (p *perses) dashboards() ([]client.Object, error) {
	dashboards, err := p.loadDashboards()
	if err != nil {
		return nil, err
	}

	var objs []client.Object
	for name, config := range dashboards {
		objs = append(objs, &persesv1alpha2.PersesDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: p.namespace,
				Labels:    p.getLabels(),
			},
			Spec: persesv1alpha2.PersesDashboardSpec{
				Config:           config,
				InstanceSelector: p.instanceSelector(),
			},
		})
	}

	return objs, nil
}

// loadDashboards reads the embedded dashboard specs relevant for the current role and returns them keyed by the
// PersesDashboard resource name (derived from the file name).
func (p *perses) loadDashboards() (map[string]persesv1alpha2.Dashboard, error) {
	requiredDashboards, ignorePaths := p.selectDashboards()

	dashboards := map[string]persesv1alpha2.Dashboard{}
	for dashboardPath, dashboardEmbed := range requiredDashboards {
		if err := fs.WalkDir(dashboardEmbed, dashboardPath, func(path string, dirEntry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if dirEntry.IsDir() {
				if ignorePaths.HasAny(strings.Split(path, "/")...) {
					return fs.SkipDir
				}
				return nil
			}

			data, err := dashboardEmbed.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading %s: %w", path, err)
			}

			dashboard, err := decodeDashboard(data)
			if err != nil {
				return fmt.Errorf("error decoding %s: %w", path, err)
			}

			dashboards[dashboardName(path)] = dashboard
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return dashboards, nil
}

func decodeDashboard(data []byte) (persesv1alpha2.Dashboard, error) {
	var dashboard persesv1alpha2.Dashboard
	err := yaml.Unmarshal(data, &dashboard)
	return dashboard, err
}

// selectDashboards returns the set of embedded dashboard directories to deploy for the current role, together with the
// path segments that should be ignored (e.g. istio or vpa dashboards which are only conditionally deployed). Perses is
// only ever deployed to the garden runtime cluster and to seeds, so only these two roles are handled.
func (p *perses) selectDashboards() (map[string]embed.FS, sets.Set[string]) {
	ignorePaths := sets.New[string]()

	if p.values.IsGardenCluster {
		requiredDashboards := map[string]embed.FS{
			gardenDashboardsPath:         gardenDashboards,
			gardenAndSeedDashboardsPath:  gardenAndSeedDashboards,
			gardenAndShootDashboardsPath: gardenAndShootDashboards,
		}
		if p.values.VPAEnabled {
			requiredDashboards[commonVpaDashboardsPath] = commonDashboards
		}
		// The garden runtime cluster does not deploy the istio dashboards.
		ignorePaths.Insert("istio")
		return requiredDashboards, ignorePaths
	}

	requiredDashboards := map[string]embed.FS{
		seedDashboardsPath:   seedDashboards,
		commonDashboardsPath: commonDashboards,
	}
	// If the seed is the garden cluster, the garden-seed dashboards are already deployed by gardener-operator, so the
	// gardenlet does not need to deploy them again.
	if !p.values.OnlyDeployDatasourcesAndDashboards {
		requiredDashboards[gardenAndSeedDashboardsPath] = gardenAndSeedDashboards
	}
	if !p.values.IncludeIstioDashboards {
		ignorePaths.Insert("istio")
	}
	if !p.values.VPAEnabled {
		ignorePaths.Insert("vpa")
	}
	return requiredDashboards, ignorePaths
}

var invalidNameChars = regexp.MustCompile(`[^a-z0-9.-]+`)

// dashboardName derives a valid DNS-1123 subdomain PersesDashboard resource name from a dashboard file path.
func dashboardName(path string) string {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.TrimSuffix(name, "-dashboard")
	name = invalidNameChars.ReplaceAllString(strings.ToLower(name), "-")
	return strings.Trim(name, "-")
}

func init() {
	// Fail fast at build/test time if any embedded dashboard produces an invalid resource name or cannot be decoded.
	for _, dashboardEmbed := range []embed.FS{gardenDashboards, seedDashboards, gardenAndSeedDashboards, gardenAndShootDashboards, commonDashboards} {
		if err := fs.WalkDir(dashboardEmbed, "dashboards", func(path string, dirEntry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if dirEntry.IsDir() {
				return nil
			}
			if errs := apivalidation.IsDNS1123Subdomain(dashboardName(path)); len(errs) > 0 {
				panic(fmt.Sprintf("embedded dashboard %s produces invalid resource name %q: %v", path, dashboardName(path), errs))
			}
			data, err := dashboardEmbed.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := decodeDashboard(data); err != nil {
				panic(fmt.Sprintf("embedded dashboard %s cannot be decoded: %v", path, err))
			}
			return nil
		}); err != nil {
			panic(err)
		}
	}
}
