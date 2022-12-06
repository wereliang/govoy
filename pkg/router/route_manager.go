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

package router

import (
	"sync"

	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/wereliang/govoy/pkg/api"
)

func NewRouteConfigManager(configs []*envoy_config_route_v3.RouteConfiguration) (api.RouteConfigManager, error) {
	rcm := &routeConfigManager{}
	for _, c := range configs {
		rcm.AddOrUpdateRouteConfig(api.ROUTE_CONFIG_STATIC, c)
	}
	return rcm, nil
}

func newRouteConfig(typ api.RouteConfigType, rc *envoy_config_route_v3.RouteConfiguration) api.RouteConfig {
	return &routeConfig{typ: typ, rc: rc}
}

type routeConfig struct {
	rc  *envoy_config_route_v3.RouteConfiguration
	typ api.RouteConfigType
}

func (rc *routeConfig) Type() api.RouteConfigType {
	return rc.typ
}

func (rc *routeConfig) Config() *envoy_config_route_v3.RouteConfiguration {
	return rc.rc
}

type routeConfigManager struct {
	routeConfigMap sync.Map
}

func (m *routeConfigManager) AddOrUpdateRouteConfig(typ api.RouteConfigType, rc *envoy_config_route_v3.RouteConfiguration) error {
	routeConfig := newRouteConfig(typ, rc)
	m.routeConfigMap.Store(rc.Name, api.NewObjectConfig(routeConfig, rc))
	return nil
}

func (m *routeConfigManager) GetRouteConfig(name string) api.RouteConfig {
	if r, ok := m.routeConfigMap.Load(name); ok {
		return r.(api.ObjectConfig).Object().(api.RouteConfig)
	}
	return nil
}

func (m *routeConfigManager) Range(cb func(string, api.ObjectConfig) bool) {
	m.routeConfigMap.Range(func(k, v interface{}) bool {
		return cb(k.(string), v.(api.ObjectConfig))
	})
}

func (m *routeConfigManager) DeleteRouteConfig(name string) error {
	m.routeConfigMap.Delete(name)
	return nil
}
