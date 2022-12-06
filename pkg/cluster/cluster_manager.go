/*
 * MIT License
 *
 * Copyright (c) 2022 wereliang
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package cluster

import (
	"fmt"
	"sync"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/log"
)

func NewClusterManager(clusters []*envoy_config_cluster_v3.Cluster) (api.ClusterManager, error) {
	cm := &clusterManager{}
	for _, cluster := range clusters {
		err := cm.AddOrUpdateCluster(cluster)
		if err != nil {
			log.Error("add cluster error:%s", err)
		}
	}
	return cm, nil
}

type clusterManager struct {
	clusterMap sync.Map
}

func (cm *clusterManager) AddOrUpdateCluster(c *envoy_config_cluster_v3.Cluster) error {

	if cluster := cm.GetCluster(c.Name); cluster != nil {
		log.Debug("cluster %s close", c.Name)
		cluster.Close()
	}

	newCluster, err := NewCluster(c)
	if err != nil {
		return err
	}
	cm.clusterMap.Store(c.Name, api.NewObjectConfig(newCluster, c))
	return nil
}

func (cm *clusterManager) GetCluster(name string) api.Cluster {
	if c, ok := cm.clusterMap.Load(name); ok {
		return c.(api.ObjectConfig).Object().(api.Cluster)
	}
	return nil
}

func (cm *clusterManager) Range(cb func(string, api.ObjectConfig) bool) {
	cm.clusterMap.Range(func(k, v interface{}) bool {
		return cb(k.(string), v.(api.ObjectConfig))
	})
}

func (cm *clusterManager) DeleteCluster(name string) error {
	cm.clusterMap.Delete(name)
	return nil
}

func (cm *clusterManager) UpdateClusterHosts(clusterName string, hosts api.HostSet) error {
	cluster := cm.GetCluster(clusterName)
	if cluster == nil {
		return fmt.Errorf("update cluster hosts fail. not found cluster:%s", clusterName)
	}
	cluster.UpdateHosts(hosts)
	return nil
}
