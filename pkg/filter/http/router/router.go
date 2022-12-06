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

package httprouter

import (
	"net"

	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_extensions_filters_http_router_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"

	"github.com/golang/protobuf/proto"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/filter"
	"github.com/wereliang/govoy/pkg/http"
	"github.com/wereliang/govoy/pkg/log"
)

func init() {
	filter.HTTPFilterFactory.Regist(new(RouterFactory))
}

type Router struct {
	cb      api.DecoderFilterCallbacks
	context api.FactoryContext
}

func (r *Router) SetDecoderFilterCallbacks(cb api.DecoderFilterCallbacks) {
	r.cb = cb
}

func (r *Router) Decode(ctx api.StreamContext) api.FilterStatus {
	route := r.cb.Route()
	entry := route.Match(ctx.Request().Header())
	if entry == nil {
		log.Error("route match fail")
		return api.Stop
	}

	log.Debug("[Cluster: %s]", entry.ClusterName())

	cluster := r.context.ClusterManager().GetCluster(entry.ClusterName())
	if cluster == nil {
		log.Error("not found cluster:%s", entry.ClusterName())
		return api.Stop
	}

	snapShot := cluster.Snapshot()
	if snapShot == nil {
		log.Error("snapshot is nil for cluster(%s)", entry.ClusterName())
		return api.Stop
	}

	lb := snapShot.LoadBalancer()
	if lb == nil {
		log.Error("loadbalancer is nil. %s", entry.ClusterName())
		return api.Stop
	}

	host := snapShot.LoadBalancer().Select(r.cb)
	log.Debug("[Endpoint: %s]", "http://"+host.Address().String())

	ctx.Request().SetHost(host.Address().String())

	err := http.Call(ctx, r.getSourceAddr(cluster.Snapshot().ClusterInfo().Config()))
	if err != nil {
		log.Error("http call error: %s", err)
		return api.Stop
	}

	return api.Continue
}

func (r *Router) getSourceAddr(cluster *envoy_config_cluster_v3.Cluster) net.Addr {
	if bind := cluster.GetUpstreamBindConfig(); bind != nil {
		if addr := bind.GetSourceAddress(); addr != nil {
			return &net.TCPAddr{IP: net.ParseIP(addr.GetAddress()), Port: int(addr.GetPortValue())}
		}
	}
	return nil
}

func (r *Router) Encode(ctx api.StreamContext) api.FilterStatus {
	return api.Continue
}

type RouterFactory struct {
}

func (f *RouterFactory) Name() string {
	return filter.HTTP_Router
}

func (f *RouterFactory) CreateEmptyConfigProto() proto.Message {
	return &envoy_extensions_filters_http_router_v3.Router{}
}

func (f *RouterFactory) CreateFilterFactory(pb proto.Message, context api.FactoryContext) api.HTTPFilterCreator {
	return func(cb api.HTTPFilterManager) {
		router := &Router{context: context}
		cb.AddDecodeFilter(router)
		cb.AddEncodeFilter(router)
	}
}
