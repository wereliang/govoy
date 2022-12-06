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
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type RouteConfigType int32

const (
	ROUTE_CONFIG_STATIC RouteConfigType = 0
	ROUTE_CONFIG_EDS    RouteConfigType = 3
)

type RouteEntry interface {
	ClusterName() string
}

type RouteConfigMatcher interface {
	Config() *envoy_config_route_v3.RouteConfiguration
	Match(RequestHeader) RouteEntry
}

type RouteConfig interface {
	Config() *envoy_config_route_v3.RouteConfiguration
	Type() RouteConfigType
}

type RouteConfigManager interface {
	GetRouteConfig(string) RouteConfig
	AddOrUpdateRouteConfig(RouteConfigType, *envoy_config_route_v3.RouteConfiguration) error
	DeleteRouteConfig(string) error
	Range(func(string, ObjectConfig) bool)
}
