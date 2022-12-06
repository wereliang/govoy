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

package hcm

import (
	"bytes"
	"fmt"

	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_filters_network_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/golang/protobuf/proto"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/filter"
	"github.com/wereliang/govoy/pkg/http"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/router"
)

func init() {
	filter.NetworkFilterFactory.Regist(new(HttpConnectionManagerFactory))
}

type HttpConnectionManager struct {
	api.ReadFilter
	config       *envoy_filters_network_v3.HttpConnectionManager
	streamServer http.StreamServer
}

func getRouteConfiguration(
	hcm *envoy_filters_network_v3.HttpConnectionManager,
	rcm api.RouteConfigManager) *envoy_config_route_v3.RouteConfiguration {

	if c := hcm.GetRouteConfig(); c != nil {
		return c
	}
	if hcm.GetRds() != nil {
		if rc := rcm.GetRouteConfig(hcm.GetRds().RouteConfigName); rc != nil {
			return rc.Config()
		}
		log.Error("not found rds route config [%s]", hcm.GetRds().RouteConfigName)
		return nil
	}
	// Not support other type now
	panic("invalid route config")
}

func newHttpConnectionManager(pb proto.Message, cb api.ConnectionCallbacks, context api.FactoryContext) api.ReadFilter {
	config := pb.(*envoy_filters_network_v3.HttpConnectionManager)
	hcm := &HttpConnectionManager{config: config}

	// TODO: 复用

	rc := getRouteConfiguration(config, context.RouteConfigManager())
	if rc == nil {
		return nil
	}
	log.Debug("[RouteConfig: %s]", rc.GetName())

	matcher := router.NewRouterMatcher(rc)
	handler := http.NewHandler(matcher, cb)

	for _, f := range hcm.config.HttpFilters {
		factory, pb := filter.GetHTTPFactory(f.GetTypedConfig(), f.Name)
		if factory == nil {
			if filter.IsWellknowName(f.Name) {
				panic(fmt.Errorf("not found factory:%s", f.Name))
			} else {
				log.Error("not support http filter: %s", f.Name)
				continue
			}
		}
		factory.CreateFilterFactory(pb, context)(handler)
	}
	hcm.streamServer = http.NewStreamServer(handler, cb)
	return hcm
}

func (f *HttpConnectionManager) OnData(buffer *bytes.Buffer) api.FilterStatus {
	if err := f.streamServer.Dispatch(buffer); err != nil {
		return api.Stop
	}
	return api.Continue
}

func (f *HttpConnectionManager) OnNewConnection() api.FilterStatus {
	return api.Continue
}

type HttpConnectionManagerFactory struct {
}

func (f *HttpConnectionManagerFactory) Name() string {
	return filter.Network_HttpConnectionManager
}

func (f *HttpConnectionManagerFactory) CreateEmptyConfigProto() proto.Message {
	return &envoy_filters_network_v3.HttpConnectionManager{}
}

func (f *HttpConnectionManagerFactory) CreateFilterFactory(
	pb proto.Message, context api.FactoryContext) api.NetworkFilterCreator {

	return func(fm api.FilterManager, cb api.ConnectionCallbacks) error {
		hcm := newHttpConnectionManager(pb, cb, context)
		if hcm == nil {
			return fmt.Errorf("create http connection manager fail")
		}
		fm.AddReadFilter(hcm)
		return nil
	}
}
