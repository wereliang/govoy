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
	"context"
	"fmt"
	"time"

	envoy_config_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/cluster"
	"github.com/wereliang/govoy/pkg/log"
	xds_v3 "github.com/wzshiming/xds/v3"
)

type XDS interface {
}

func NewXDS(cfg *envoy_config_bootstrap_v3.Bootstrap, ctx api.FactoryContext) (XDS, error) {

	xds := &xdsImpl{cfg: cfg, context: ctx}
	var dynamic *envoy_config_bootstrap_v3.Bootstrap_DynamicResources

	if dynamic = cfg.GetDynamicResources(); dynamic == nil {
		log.Debug("no dynamic resources")
		return nil, nil
	}
	if err := xds.check(dynamic); err != nil {
		return nil, err
	}
	if err := xds.Start(); err != nil {
		return nil, err
	}
	return xds, nil
}

type xdsImpl struct {
	cfg         *envoy_config_bootstrap_v3.Bootstrap
	context     api.FactoryContext
	grpcCluster string
}

func (x *xdsImpl) check(dynamic *envoy_config_bootstrap_v3.Bootstrap_DynamicResources) error {

	log.Debug("dynamic config: %#v", dynamic)

	// just support ads
	ads := dynamic.GetAdsConfig()
	if ads == nil {
		return fmt.Errorf("ads no config")
	}
	log.Debug("ads config: %#v", ads)

	if ads.GetApiType() != envoy_config_core_v3.ApiConfigSource_GRPC {
		return fmt.Errorf("not support ads api type: %v", ads.GetApiType())
	}

	if ads.GetGrpcServices() == nil {
		return fmt.Errorf("grpc service is nil")
	}

	if ads.GetGrpcServices()[0].GetEnvoyGrpc() == nil {
		return fmt.Errorf("envoy grpc is nil")
	}

	x.grpcCluster = ads.GetGrpcServices()[0].GetEnvoyGrpc().GetClusterName()
	return nil
}

func (x *xdsImpl) Start() error {
	handler := &handler{context: x.context}
	cli, err := x.newClient(handler)
	if err != nil {
		return err
	}
	handler.SetClient(&xdsClient{cli})

	x.run(cli)
	return nil
}

func (x *xdsImpl) run(cli *xds_v3.Client) {
	go func() {
		for {
			log.Debug("xds client running...")
			err := cli.Run(context.Background())
			if err != nil {
				log.Error("xds run error: %s", err)
			}
			time.Sleep(time.Second * 5)
		}
	}()
}

func (x *xdsImpl) newClient(handler XDSHandler) (*xds_v3.Client, error) {
	conf := xds_v3.Config{}
	conf.HandleCDS = func(cli *xds_v3.Client, clusters []*envoy_config_cluster_v3.Cluster) {
		handler.HandleCDS(clusters)
	}
	conf.HandleLDS = func(cli *xds_v3.Client, listeners []*envoy_config_listener_v3.Listener) {
		handler.HandleLDS(listeners)
	}
	conf.HandleRDS = func(cli *xds_v3.Client, routes []*envoy_config_route_v3.RouteConfiguration) {
		handler.HandleRDS(routes)
	}
	conf.HandleEDS = func(cli *xds_v3.Client, endpoints []*envoy_config_endpoint_v3.ClusterLoadAssignment) {
		handler.HandleEDS(endpoints)
	}
	conf.OnConnect = func(cli *xds_v3.Client) error {
		return handler.OnConnect()
	}

	node := x.cfg.GetNode()
	if node != nil {
		conf.NodeConfig.NodeID = node.GetId()
		conf.NodeConfig.Metadata = node.Metadata.AsMap()
	}

	addr, err := x.getClusterAddr()
	if err != nil {
		return nil, err
	}

	return xds_v3.NewClient(addr, nil, &conf), nil
}

func (x *xdsImpl) getClusterAddr() (string, error) {
	c := x.context.ClusterManager().GetCluster(x.grpcCluster)
	if c == nil {
		return "", fmt.Errorf("xds grpc cluster(%s) is nil", x.grpcCluster)
	}
	hosts, err := cluster.GetClusterEndpoint(c.Snapshot().ClusterInfo().Config())
	if err != nil {
		return "", err
	}
	return hosts[0].Address().String(), nil
}
