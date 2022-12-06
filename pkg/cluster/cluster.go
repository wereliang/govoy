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
	"net"
	"sync/atomic"
	"time"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/lb"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/utils"
)

type ClusterCreator func(*envoy_config_cluster_v3.Cluster) (api.Cluster, error)

var (
	clusterFactory = make(map[api.ClusterType]ClusterCreator)
)

func registCluster(t api.ClusterType, creator ClusterCreator) {
	clusterFactory[t] = creator
}

func NewCluster(c *envoy_config_cluster_v3.Cluster) (api.Cluster, error) {
	t := api.ClusterType(c.GetType())
	if creator, ok := clusterFactory[t]; ok {
		return creator(c)
	}
	return nil, fmt.Errorf("not support cluster type: %v", c.GetType())
}

func GetClusterEndpoint(c *envoy_config_cluster_v3.Cluster) (api.HostSet, error) {
	la := c.GetLoadAssignment()
	if la == nil {
		return nil, fmt.Errorf("nil load assignment")
	}
	return GetEndpointFromClusterLoad(c.GetType(), la)
}

func GetEndpointFromClusterLoad(
	typ envoy_config_cluster_v3.Cluster_DiscoveryType,
	la *envoy_config_endpoint_v3.ClusterLoadAssignment) (api.HostSet, error) {

	var hostSet api.HostSet
	for _, ledp := range la.GetEndpoints() {
		for _, lbedp := range ledp.GetLbEndpoints() {
			host, err := NewHostByEndpoint(lbedp, api.ClusterType(typ))
			if err != nil {
				return nil, err
			}
			hostSet = append(hostSet, host)
		}
	}
	return hostSet, nil
}

func newSimpleCluster(cluster *envoy_config_cluster_v3.Cluster) *simpleCluster {

	clusterType := api.ClusterType(cluster.GetType())
	lbType := api.LoadBalancerType(cluster.GetLbPolicy())
	if clusterType == api.Cluster_ORIGINAL_DST {
		lbType = api.Original_Dst
	}
	log.Debug("cluster:%s type:%d lb:%d", cluster.Name, clusterType, lbType)

	return &simpleCluster{
		info: &clusterInfo{
			name:        cluster.Name,
			clusterType: clusterType,
			lbType:      lbType,
			config:      cluster,
			ts:          time.Now(),
		},
	}
}

type simpleCluster struct {
	snapShot atomic.Value
	info     api.ClusterInfo
}

func (c *simpleCluster) Snapshot() api.ClusterSnapshot {
	ss := c.snapShot.Load()
	if css, ok := ss.(*clusterSnapShot); ok {
		return css
	}
	return nil
}

func (c *simpleCluster) UpdateHosts(hosts []api.Host) {
	snapShot := &clusterSnapShot{
		clusterInfo: c.info,
		lb:          lb.NewLoadBalancer(c.info.LbType(), hosts),
		hosts:       hosts,
	}
	c.snapShot.Store(snapShot)
}

func (c *simpleCluster) getConfigHosts() (api.HostSet, error) {
	return GetClusterEndpoint(c.info.Config())
}

func (c *simpleCluster) Close() {}

type clusterSnapShot struct {
	clusterInfo api.ClusterInfo
	lb          api.LoadBalancer
	hosts       api.HostSet
}

func (cs *clusterSnapShot) ClusterInfo() api.ClusterInfo {
	return cs.clusterInfo
}

func (cs *clusterSnapShot) HostSet() api.HostSet {
	return cs.hosts
}

func (cs *clusterSnapShot) LoadBalancer() api.LoadBalancer {
	return cs.lb
}

type clusterInfo struct {
	name        string
	clusterType api.ClusterType
	lbType      api.LoadBalancerType
	config      *envoy_config_cluster_v3.Cluster
	ts          time.Time
}

func (c *clusterInfo) Name() string {
	return c.name
}

func (c *clusterInfo) ClusterType() api.ClusterType {
	return c.clusterType
}

func (c *clusterInfo) LbType() api.LoadBalancerType {
	return c.lbType
}

func (c *clusterInfo) Config() *envoy_config_cluster_v3.Cluster {
	return c.config
}

func NewHostByEndpoint(lbedp *envoy_config_endpoint_v3.LbEndpoint, ctype api.ClusterType) (api.Host, error) {
	edp := lbedp.GetEndpoint()
	if edp == nil {
		return nil, fmt.Errorf("nil endpoint")
	}

	var (
		addr net.Addr
		err  error
	)
	switch ctype {
	case api.Cluster_Static, api.Cluster_EDS:
		addr, err = utils.ToNetAddr(edp.GetAddress())
	case api.Cluster_Strict_DNS:
		addr, err = utils.ToDNSAddr(edp.GetAddress())
	default:
		err = fmt.Errorf("not support clusterType(%d) for new host", ctype)
	}
	if err != nil {
		return nil, err
	}

	h := api.NewHost(addr)
	if lbedp.GetLoadBalancingWeight() != nil {
		h.SetWeight(lbedp.GetLoadBalancingWeight().GetValue())
	} else {
		// set default weight
		h.SetWeight(api.DEFAULT_WEIGHT)
	}
	return h, nil
}
