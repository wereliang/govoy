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
	"fmt"

	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/fanyang01/radix"
	"github.com/wereliang/govoy/pkg/api"
)

// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto
// Domain search order:
// Exact domain names: www.foo.com.
// Suffix domain wildcards: *.foo.com or *-bar.foo.com.
// Prefix domain wildcards: foo.* or foo-*.
// Special wildcard * matching any domain.

type routeEntry struct {
	cluster string
}

func (re *routeEntry) ClusterName() string {
	return re.cluster
}

func NewRouteEntry(cluster string) api.RouteEntry {
	return &routeEntry{cluster}
}

func NewRouterMatcher(c *envoy_config_route_v3.RouteConfiguration) api.RouteConfigMatcher {
	rc := &routeConfigMatcher{config: c, domains: radix.NewPatternTrie()}
	rc.build()
	return rc
}

type routeConfigMatcher struct {
	config  *envoy_config_route_v3.RouteConfiguration
	domains *radix.PatternTrie
}

func (rc *routeConfigMatcher) build() {
	for _, vhConfig := range rc.config.VirtualHosts {

		vh := &virtualHost{
			name:   vhConfig.Name,
			routes: NewRouter()}

		for _, route := range vhConfig.Routes {
			var r Route
			if obj := route.Match.GetPath(); obj != "" {
				r = vh.routes.Path(obj)
			} else if obj := route.Match.GetPrefix(); obj != "" {
				r = vh.routes.PathPrefix(obj)
			} else {
				panic(fmt.Sprintf("invalid match path:%#v", route))
			}
			// just support cluster action
			if cluster := route.GetRoute().GetCluster(); cluster != "" {
				r.Handler(cluster)
			} else {
				panic("invalid route action(just support cluster)")
			}
		}

		for _, host := range vhConfig.Domains {
			rc.domains.Add(host, vh)
		}
	}
}

func (rc *routeConfigMatcher) Config() *envoy_config_route_v3.RouteConfiguration {
	return rc.config
}

func (rc *routeConfigMatcher) Match(header api.RequestHeader) api.RouteEntry {
	v, ok := rc.domains.Lookup(string(header.Host()))
	if !ok {
		return nil
	}
	if re := v.(*virtualHost).Match(header); re != nil {
		return re
	}
	return nil
}

type virtualHost struct {
	name string
	// domains Router
	routes Router
}

func (vh *virtualHost) Match(header api.RequestHeader) api.RouteEntry {
	// if _, b := vh.domains.Match(header); !b {
	// 	return nil
	// }

	if h, b := vh.routes.Match(header); !b {
		return nil
	} else {
		return NewRouteEntry(h.(string))
	}
}
