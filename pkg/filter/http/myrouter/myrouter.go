package myrouter

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/router"
)

var defaultRouter = &myRouterExecuter{}

type myRouterExecuter struct {
}

func (r *myRouterExecuter) Config() *envoy_config_route_v3.RouteConfiguration {
	return nil
}

func (r *myRouterExecuter) Match(header api.RequestHeader) api.RouteEntry {
	if string(header.Path()) != "/mesh" {
		return nil
	}
	if appid := header.Get("Dst-Appid"); appid == nil {
		return nil
	} else {
		return router.NewRouteEntry(string(appid))
	}
}
