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

package server

import (
	"fmt"

	_ "net/http/pprof"

	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/wereliang/govoy/pkg/admin"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/cluster"
	"github.com/wereliang/govoy/pkg/config"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/router"
	"github.com/wereliang/govoy/pkg/xds"
)

type Govoy struct {
	cfg config.GovoyConfig
	am  admin.Admin
	lm  api.ListenerManager
	cm  api.ClusterManager
	rm  api.RouteConfigManager
	xds xds.XDS
}

func NewGovoy(path string) (*Govoy, error) {
	govoy := &Govoy{}
	err := govoy.init(path)
	if err != nil {
		return nil, err
	}
	return govoy, nil
}

func (s *Govoy) init(path string) error {
	var err error
	if err = s.loadConfig(path); err != nil {
		return err
	}
	if err = s.initAdmin(); err != nil {
		return err
	}
	if err = s.initCluster(); err != nil {
		return err
	}
	if err = s.initRouteConfig(); err != nil {
		return err
	}
	if err = s.initListener(); err != nil {
		return err
	}
	if err = s.initXDS(); err != nil {
		return err
	}
	return nil
}

func (s *Govoy) loadConfig(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("invalid config path")
	}
	cfg, err := config.LoadBootstrapConfig(path)
	if err != nil {
		return err
	}
	s.cfg = cfg

	bs := s.bootrap()
	log.Trace("config: %#v\n", bs)
	for _, l := range bs.StaticResources.Listeners {
		log.Debug("listener: %#v", l)
	}
	return nil
}

func (s *Govoy) bootrap() *bootstrapv3.Bootstrap {
	return s.cfg.Bootstrap().Config()
}

func (s *Govoy) initAdmin() error {
	if s.bootrap().Admin != nil {
		s.am = admin.NewAdmin(s.cfg, s)
	}
	return nil
}

func (s *Govoy) initCluster() error {
	var err error
	if s.cm, err = cluster.NewClusterManager(
		s.bootrap().StaticResources.Clusters); err != nil {
		return err
	}
	return nil
}

func (s *Govoy) initListener() error {
	var err error
	if s.lm, err = NewListenerManager(s); err != nil {
		return err
	}

	for _, l := range s.bootrap().StaticResources.Listeners {
		err = s.lm.AddOrUpdateListener(api.LISTENER_STATIC, l)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Govoy) initRouteConfig() error {
	s.rm, _ = router.NewRouteConfigManager(nil)
	return nil
}

func (s *Govoy) initXDS() error {
	var err error
	if s.xds, err = xds.NewXDS(s.bootrap(), s); err != nil {
		return err
	}
	return nil
}

func (s *Govoy) ClusterManager() api.ClusterManager {
	return s.cm
}

func (s *Govoy) RouteConfigManager() api.RouteConfigManager {
	return s.rm
}

func (s *Govoy) ListenerManager() api.ListenerManager {
	return s.lm
}

func (s *Govoy) Start() error {
	// admin
	if s.am != nil {
		if err := s.am.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Govoy) Stop() {
	// TODO
}
