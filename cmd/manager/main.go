/*
Copyright 2018 The Multicluster-Controller Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"openeoe/openeoe/omcplog"
	"openeoe/openeoe/openeoe-migration/pkg/controller/migration"
	"openeoe/openeoe/util/clusterManager"
	"openeoe/openeoe/util/controller/logLevel"
	"openeoe/openeoe/util/controller/reshape"

	"admiralty.io/multicluster-controller/pkg/cluster"
	"admiralty.io/multicluster-controller/pkg/manager"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	logLevel.KetiLogInit()

	for {
		cm := clusterManager.NewClusterManager()
		host_ctx := "openeoe"
		namespace := "openeoe"

		host_cfg := cm.Host_config

		live := cluster.New(host_ctx, host_cfg, cluster.Options{CacheOptions: cluster.CacheOptions{}})

		ghosts := []*cluster.Cluster{}
		for _, ghost_cluster := range cm.Cluster_list.Items {
			ghost_ctx := ghost_cluster.Name
			ghost_cfg := cm.Cluster_configs[ghost_ctx]

			ghost := cluster.New(ghost_ctx, ghost_cfg, cluster.Options{CacheOptions: cluster.CacheOptions{}})

			ghosts = append(ghosts, ghost)
		}
		co, _ := migration.NewController(live, ghosts, namespace, cm)
		loglevel_cont, _ := logLevel.NewController(live, ghosts, namespace)
		m := manager.New()
		m.AddController(co)
		m.AddController(loglevel_cont)
		stop := reshape.SetupSignalHandler()
		if err := m.Start(stop); err != nil {
			omcplog.V(4).Info("[OpenEOE] error:  ", err)
		}

	}
}
