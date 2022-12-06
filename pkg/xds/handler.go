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

package xds

import (
	"sort"
	"strings"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/cluster"
	"github.com/wereliang/govoy/pkg/log"
	xds_v3 "github.com/wzshiming/xds/v3"
)

type XDSHandler interface {
	HandleCDS(clusters []*envoy_config_cluster_v3.Cluster)
	HandleEDS(endpoints []*envoy_config_endpoint_v3.ClusterLoadAssignment)
	HandleLDS(listeners []*envoy_config_listener_v3.Listener)
	HandleRDS(routes []*envoy_config_route_v3.RouteConfiguration)
	HandleSDS(secrets []*envoy_extensions_transport_sockets_tls_v3.Secret)
	OnConnect() error
	SetClient(XDSClient)
}

type handler struct {
	cli       XDSClient
	clusters  []string
	listeners []string
	routes    []string
	context   api.FactoryContext
}

func (h *handler) SetClient(cli XDSClient) {
	h.cli = cli
}

func (h *handler) OnConnect() error {
	var err error
	if err = h.cli.SendCDS(nil); err != nil {
		log.Error("xds send cds error. %s", err)
		return err
	}
	log.Debug("Request CDS OnConnect")
	if err = h.cli.SendLDS(nil); err != nil {
		log.Error("xds send lds error. %s", err)
		return err
	}
	log.Debug("Request LDS OnConnect")
	return nil
}

func (h *handler) deleteXDS(olds, news []string, deletor func(string)) []string {
	if len(olds) == 0 {
		return nil
	}
	var (
		dels   []string
		oi, ni int
		olen   = len(olds)
	)

	for ni < len(news) {
		name := news[ni]
		oname := olds[oi]
		if name == oname {
			oi++
			ni++
		} else if name > oname {
			dels = append(dels, oname)
			oi++
		} else {
			ni++
		}
		if oi == olen {
			break
		}
	}

	if oi < olen {
		dels = append(dels, olds[oi:]...)
	}

	for _, d := range dels {
		deletor(d)
	}
	return dels
}

func (h *handler) HandleCDS(clusters []*envoy_config_cluster_v3.Cluster) {

	log.Debug("Response CDS: %d", len(clusters))
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})

	var endpoints, newClusters []string
	for _, cluster := range clusters {
		err := h.context.ClusterManager().AddOrUpdateCluster(cluster)
		if err != nil {
			log.Error("AddOrUpdateCluster(%s) error: %s", cluster.GetName(), err)
		} else {
			log.Debug("add or update cluster: %s", cluster.Name)
		}
		endpoints = append(endpoints, xds_v3.GetEndpointNames(cluster)...)
		newClusters = append(newClusters, cluster.Name)
	}

	dels := h.deleteXDS(h.clusters, newClusters, func(s string) {
		h.context.ClusterManager().DeleteCluster(s)
	})
	log.Debug("delete cluster: %#v", dels)

	h.clusters = newClusters

	log.Debug("Request EDS. %d %s", len(endpoints), strings.Join(endpoints, ","))
	h.cli.SendEDS(endpoints)
}

func (h *handler) HandleEDS(endpoints []*envoy_config_endpoint_v3.ClusterLoadAssignment) {
	log.Debug("Response EDS: %d", len(endpoints))
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].ClusterName < endpoints[j].ClusterName
	})
	for _, endpoint := range endpoints {
		hosts, err := cluster.GetEndpointFromClusterLoad(envoy_config_cluster_v3.Cluster_EDS, endpoint)
		if err != nil {
			log.Error("get host fail. %s", err)
			continue
		}

		err = h.context.ClusterManager().UpdateClusterHosts(endpoint.GetClusterName(), hosts)
		if err != nil {
			log.Error("update cluster[%s] hosts error. %s", endpoint.GetClusterName(), err)
		} else {
			log.Debug("update cluster[%s] hosts success", endpoint.GetClusterName())
		}
	}
}

func (h *handler) HandleLDS(listeners []*envoy_config_listener_v3.Listener) {
	log.Debug("Response LDS: %d", len(listeners))
	sort.Slice(listeners, func(i, j int) bool {
		return listeners[i].Name < listeners[j].Name
	})

	var routes, newListeners []string
	for _, listener := range listeners {
		err := h.context.ListenerManager().AddOrUpdateListener(api.LISTENER_EDS, listener)
		if err != nil {
			log.Error("add or update listener(%s) error: %s", listener.GetName(), err)
		} else {
			log.Debug("add or update listener: %s", listener.GetName())
		}

		routes = append(routes, xds_v3.GetRouteNames(listener)...)
		newListeners = append(newListeners, listener.Name)
	}

	dels := h.deleteXDS(h.listeners, newListeners, func(s string) {
		h.context.ListenerManager().DeleteListener(s)
	})
	log.Debug("delete listener: %#v", dels)

	h.listeners = newListeners

	log.Debug("Request RDS. %d %s", len(routes), strings.Join(routes, ","))
	h.cli.SendRDS(routes)
}

func (h *handler) HandleRDS(routes []*envoy_config_route_v3.RouteConfiguration) {
	log.Debug("Response RDS: %d", len(routes))
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Name < routes[j].Name
	})

	var newRoutes []string
	for _, route := range routes {
		log.Debug("route: %s", route.Name)
		err := h.context.RouteConfigManager().AddOrUpdateRouteConfig(api.ROUTE_CONFIG_EDS, route)
		if err != nil {
			log.Error("add or update route(%s) error: %s", route.GetName(), err)
		} else {
			log.Debug("add or update route: %s", route.GetName())
		}
		newRoutes = append(newRoutes, route.GetName())
	}

	dels := h.deleteXDS(h.routes, newRoutes, func(s string) {
		h.context.RouteConfigManager().DeleteRouteConfig(s)
	})
	log.Debug("delete routes: %#v", dels)
	h.routes = newRoutes
}

func (h *handler) HandleSDS(secrets []*envoy_extensions_transport_sockets_tls_v3.Secret) {

}
