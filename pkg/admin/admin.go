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

package admin

import (
	"encoding/json"
	"fmt"
	"net/http"

	envoy_admin_v3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	envoy_config_bootstrap_v3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/config"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/utils"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewAdmin(cfg config.GovoyConfig, ctx api.FactoryContext) Admin {
	return &adminServer{
		govoyConfig: cfg,
		adminConfig: cfg.Bootstrap().Config().Admin,
		context:     ctx,
	}
}

type Admin interface {
	Start() error
	Config() *envoy_config_bootstrap_v3.Admin
}

type adminServer struct {
	govoyConfig config.GovoyConfig
	adminConfig *envoy_config_bootstrap_v3.Admin
	context     api.FactoryContext
}

func (s *adminServer) Start() error {
	addr, err := utils.ToNetAddr(s.adminConfig.Address)
	if err != nil {
		return err
	}

	s.handle()
	go func() {
		log.Info("start admin:%s", addr.String())
		if err := http.ListenAndServe(addr.String(), nil); err != nil {
			panic(err)
		}
	}()
	return nil
}

func (s *adminServer) Config() *envoy_config_bootstrap_v3.Admin {
	return s.adminConfig
}

func (s *adminServer) handle() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Govoy Admin"))
	})
	http.HandleFunc("/config_dump", s.configDump)
	http.HandleFunc("/cluster", s.cluster)
}

func (s *adminServer) configDump(w http.ResponseWriter, r *http.Request) {
	dmp := &envoy_admin_v3.ConfigDump{}
	var any *anypb.Any
	// bootstrap
	any, _ = ptypes.MarshalAny(s.dumpBootstrap())
	dmp.Configs = append(dmp.Configs, any)

	// listener
	any, _ = ptypes.MarshalAny(s.dumpListener())
	dmp.Configs = append(dmp.Configs, any)
	// routeconfig
	any, _ = ptypes.MarshalAny(s.dumpRouteConfig())
	dmp.Configs = append(dmp.Configs, any)

	any, _ = ptypes.MarshalAny(s.dumpCluster())
	dmp.Configs = append(dmp.Configs, any)

	mar := jsonpb.Marshaler{}
	mar.Marshal(w, dmp)
	log.Debug("call config dump")
}

func (s *adminServer) dumpBootstrap() *envoy_admin_v3.BootstrapConfigDump {
	return &envoy_admin_v3.BootstrapConfigDump{
		Bootstrap:   s.govoyConfig.Bootstrap().Config(),
		LastUpdated: timestamppb.New(s.govoyConfig.Bootstrap().LastUpdated()),
	}
}

func (s *adminServer) dumpListener() *envoy_admin_v3.ListenersConfigDump {
	listenerDump := &envoy_admin_v3.ListenersConfigDump{}
	listenerManager := s.context.ListenerManager()
	listenerManager.Range(func(name string, oc api.ObjectConfig) bool {
		actl := oc.Object().(api.ActiveListener)
		any, _ := ptypes.MarshalAny(oc.Config().(*envoy_config_listener_v3.Listener))
		switch actl.Type() {
		case api.LISTENER_STATIC:
			listenerDump.StaticListeners = append(listenerDump.StaticListeners,
				&envoy_admin_v3.ListenersConfigDump_StaticListener{
					Listener:    any,
					LastUpdated: timestamppb.New(oc.Updated()),
				})
		case api.LISTENER_EDS:
			listenerDump.DynamicListeners = append(listenerDump.DynamicListeners,
				&envoy_admin_v3.ListenersConfigDump_DynamicListener{
					ActiveState: &envoy_admin_v3.ListenersConfigDump_DynamicListenerState{
						Listener:    any,
						LastUpdated: timestamppb.New(oc.Updated()),
					}})
		}
		return true
	})
	return listenerDump
}

func (s *adminServer) dumpCluster() *envoy_admin_v3.ClustersConfigDump {
	clusterDump := &envoy_admin_v3.ClustersConfigDump{}
	clusterManager := s.context.ClusterManager()
	clusterManager.Range(func(name string, oc api.ObjectConfig) bool {
		info := oc.Object().(api.Cluster).Snapshot().ClusterInfo()
		any, _ := ptypes.MarshalAny(oc.Config().(*envoy_config_cluster_v3.Cluster))
		switch info.ClusterType() {
		case api.Cluster_Static:
			clusterDump.StaticClusters = append(clusterDump.StaticClusters,
				&envoy_admin_v3.ClustersConfigDump_StaticCluster{
					Cluster:     any,
					LastUpdated: timestamppb.New(oc.Updated()),
				})
		case api.Cluster_EDS:
			clusterDump.DynamicActiveClusters = append(clusterDump.DynamicActiveClusters,
				&envoy_admin_v3.ClustersConfigDump_DynamicCluster{
					Cluster:     any,
					LastUpdated: timestamppb.New(oc.Updated()),
				})
		}
		return true
	})
	return clusterDump
}

func (s *adminServer) dumpRouteConfig() *envoy_admin_v3.RoutesConfigDump {
	routeConfigDump := &envoy_admin_v3.RoutesConfigDump{}
	routeConfigManager := s.context.RouteConfigManager()
	routeConfigManager.Range(func(name string, oc api.ObjectConfig) bool {
		rc := oc.Object().(api.RouteConfig)
		any, _ := ptypes.MarshalAny(rc.Config())
		switch rc.Type() {
		case api.ROUTE_CONFIG_STATIC:
			routeConfigDump.StaticRouteConfigs = append(
				routeConfigDump.StaticRouteConfigs,
				&envoy_admin_v3.RoutesConfigDump_StaticRouteConfig{
					RouteConfig: any,
					LastUpdated: timestamppb.New(oc.Updated())})
		case api.ROUTE_CONFIG_EDS:
			routeConfigDump.DynamicRouteConfigs = append(
				routeConfigDump.DynamicRouteConfigs,
				&envoy_admin_v3.RoutesConfigDump_DynamicRouteConfig{
					RouteConfig: any,
					LastUpdated: timestamppb.New(oc.Updated())})
		}
		return true
	})
	return routeConfigDump
}

func (s *adminServer) cluster(w http.ResponseWriter, r *http.Request) {

	var err error
	r.ParseForm()

	temp := struct {
		Cluster *envoy_config_cluster_v3.Cluster `json:"cluster"`
		Hosts   api.HostSet                      `json:"hosts"`
	}{}

	if name := r.FormValue("name"); name == "" {
		err = fmt.Errorf("param error. [name]")
	} else {
		cluster := s.context.ClusterManager().GetCluster(name)
		if cluster == nil {
			err = fmt.Errorf("not found cluster: %s", name)
		} else {
			snapShot := cluster.Snapshot()
			temp.Cluster = snapShot.ClusterInfo().Config()
			temp.Hosts = snapShot.HostSet()
			for _, h := range temp.Hosts {
				log.Debug("ip:%s weight:%d", h.Address().String(), h.Weight())
			}
			data, _ := json.Marshal(temp)
			w.Write(data)
			return
		}
	}

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
}
