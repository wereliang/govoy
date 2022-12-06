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
	"net"
	"sync"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/filter"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/utils"
)

func NewListenerManager(context api.FactoryContext) (api.ListenerManager, error) {
	return &listenerManagerImpl{context: context}, nil
}

type listenerManagerImpl struct {
	listenerMap sync.Map
	context     api.FactoryContext
}

func (lm *listenerManagerImpl) AddOrUpdateListener(typ api.ListenerType, pb *envoy_config_listener_v3.Listener) error {
	var (
		actl   api.ActiveListener
		netl   api.Listener
		err    error
		update bool = false
	)

	if any, ok := lm.listenerMap.Load(pb.Name); ok {
		netl = any.(api.ObjectConfig).Object().(api.ActiveListener).Listener()
		// addr must be same
		if addr, err := utils.ToNetAddr(pb.GetAddress()); err != nil {
			return err
		} else {
			if netl.Addr().String() != addr.String() {
				return fmt.Errorf("addr not same. (%s) != (%s)",
					netl.Addr().String(), addr.String())
			}
		}
		update = true
	}

	if actl, err = NewActiveListener(typ, pb, lm.context, netl); err != nil {
		return err
	}
	if err = lm.addListenerFilter(pb.GetListenerFilters(), actl); err != nil {
		return err
	}

	actl.Listener().SetCallback(actl.(api.ListenerCallback))
	lm.listenerMap.Store(pb.Name, api.NewObjectConfig(actl, pb))

	// TODO: stop and destory
	if !update && actl.GetBindToPort() {
		go func() {
			if err := actl.Start(); err != nil {
				panic(err)
			}
		}()
	}

	if len(pb.ListenerFilters) > 0 {
		log.Debug("listener filter: %#v type:%#v",
			pb.ListenerFilters[0], pb.ListenerFilters[0].GetTypedConfig())
	}

	return nil
}

func (lm *listenerManagerImpl) addListenerFilter(
	filters []*envoy_config_listener_v3.ListenerFilter, actl api.ActiveListener) error {

	for _, f := range filters {
		factory, pb := filter.GetListenerFactory(f.GetTypedConfig(), f.Name)
		if factory == nil {
			if filter.IsWellknowName(f.Name) {
				panic(fmt.Errorf("not found listener factory:%s", f.Name))
			} else {
				log.Error("not support listener filter (%s) now", f.Name)
				continue
			}
		}
		factory.CreateFilterFactory(pb, lm.context)(actl)
	}

	// add original dst filter if use_original_dst flag set
	if actl.GetUseOriginalDst() {
		return lm.buildOriginalDstListenerFilter(actl)
	}
	return nil
}

func (lm *listenerManagerImpl) buildOriginalDstListenerFilter(actl api.ActiveListener) error {
	factory, pb := filter.GetListenerFactory(nil, filter.Listener_OriginalDst)
	if factory == nil {
		return fmt.Errorf("not found factory:%s", filter.Listener_OriginalDst)
	}
	factory.CreateFilterFactory(pb, lm.context)(actl)
	return nil
}

// TODO: 优化
func (lm *listenerManagerImpl) FindListenerByAddress(addr net.Addr) api.ActiveListener {
	var actl api.ActiveListener
	lm.listenerMap.Range(func(k, v interface{}) bool {
		l := v.(api.ObjectConfig).Object().(api.ActiveListener)
		a := l.Listener().Addr()
		if addr.Network() != a.Network() {
			return true
		}
		tcpDst := a.(*net.TCPAddr)
		tcpSrc := addr.(*net.TCPAddr)
		if (tcpDst.IP.String() == "0.0.0.0" && tcpDst.Port == tcpSrc.Port) ||
			(tcpDst.String() == tcpSrc.String()) {
			actl = l
			return false
		}
		return true
	})
	return actl
}

func (lm *listenerManagerImpl) FindListenerByName(name string) api.ActiveListener {
	if any, ok := lm.listenerMap.Load(name); ok {
		return any.(api.ObjectConfig).Object().(api.ActiveListener)
	}
	return nil
}

func (lm *listenerManagerImpl) Range(cb func(string, api.ObjectConfig) bool) {
	lm.listenerMap.Range(func(k, v interface{}) bool {
		return cb(k.(string), v.(api.ObjectConfig))
	})
}

func (lm *listenerManagerImpl) DeleteListener(name string) error {
	lm.listenerMap.Delete(name)
	return nil
}
