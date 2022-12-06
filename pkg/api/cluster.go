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

package api

import (
	"net"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
)

type HostInfo interface {
	SetWeight(uint32)
	Weight() uint32
}

type Host interface {
	HostInfo
	Address() net.Addr
}

type HostSet []Host

// ClusterSnapshot is a thread-safe cluster snapshot
type ClusterSnapshot interface {
	// HostSet returns the cluster snapshot's host set
	HostSet() HostSet

	// ClusterInfo returns the cluster snapshot's cluster info
	ClusterInfo() ClusterInfo

	// LoadBalancer returns the cluster snapshot's load balancer
	LoadBalancer() LoadBalancer
}

// Cluster is a group of upstream hosts
type Cluster interface {
	// Snapshot returns the cluster snapshot, which contains cluster info, hostset and load balancer
	Snapshot() ClusterSnapshot

	// UpdateHosts updates the host set's hosts
	UpdateHosts([]Host)

	// Close destory cluster
	Close()
}

// ClusterManager is a manager for cluster
type ClusterManager interface {
	// AddOrUpdateCluster add or update cluster
	AddOrUpdateCluster(*envoy_config_cluster_v3.Cluster) error

	// DeleteCluster delete cluster by name
	DeleteCluster(string) error

	// GetCluster get cluster by name
	GetCluster(string) Cluster

	// UpdateClusterHosts update cluster by name and hosts
	UpdateClusterHosts(string, HostSet) error

	// Range liken sync.Map
	Range(func(string, ObjectConfig) bool)
}

type ClusterType int32

const (
	Cluster_Static       ClusterType = 0
	Cluster_Strict_DNS   ClusterType = 1
	Cluster_Logical_DNS  ClusterType = 2
	Cluster_EDS          ClusterType = 3
	Cluster_ORIGINAL_DST ClusterType = 4
)

// ClusterInfo defines a cluster's information
type ClusterInfo interface {
	// Name returns the cluster name
	Name() string

	// ClusterType returns the cluster type
	ClusterType() ClusterType

	// LbType returns the cluster's load balancer type
	LbType() LoadBalancerType

	// Config returns cluster config
	Config() *envoy_config_cluster_v3.Cluster
}

const (
	DEFAULT_WEIGHT = 100
)

func NewHost(addr net.Addr) Host {
	return &host{addr: addr}
}

type host struct {
	weight uint32
	addr   net.Addr
}

func (h *host) Weight() uint32 {
	return h.weight
}

func (h *host) SetWeight(w uint32) {
	h.weight = w
}

func (h *host) Address() net.Addr {
	return h.addr
}
