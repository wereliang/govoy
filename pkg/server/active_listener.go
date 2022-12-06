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

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/filter"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/wereliang/govoy/pkg/network"
	"github.com/wereliang/govoy/pkg/utils"
)

func NewActiveListener(typ api.ListenerType, pb *envoy_config_listener_v3.Listener,
	context api.FactoryContext, l api.Listener) (api.ActiveListener, error) {

	if l == nil {
		addr, err := utils.ToNetAddr(pb.GetAddress())
		if err != nil {
			return nil, err
		}
		l = network.NewListener(addr)
	}

	fcm := newFilterChainManager(pb.GetFilterChains(), pb.GetDefaultFilterChain())
	al := &activeListener{
		pb:                 pb,
		context:            context,
		typ:                typ,
		listener:           l,
		filterChainManager: fcm,
		useOriginalDst:     false,
		bindToPort:         true}

	if pb.GetUseOriginalDst() != nil {
		al.useOriginalDst = pb.GetUseOriginalDst().GetValue()
	}
	if pb.GetBindToPort() != nil {
		al.bindToPort = pb.GetBindToPort().GetValue()
	}
	// l.SetCallback(al)
	return al, nil
}

type activeListener struct {
	pb                 *envoy_config_listener_v3.Listener
	filters            []api.ListenerFilter
	listener           api.Listener
	context            api.FactoryContext
	typ                api.ListenerType
	filterChainManager api.FilterChainManager
	useOriginalDst     bool
	bindToPort         bool
}

func (al *activeListener) Listener() api.Listener {
	return al.listener
}

func (al *activeListener) Type() api.ListenerType {
	return al.typ
}

func (al *activeListener) GetUseOriginalDst() bool {
	return al.useOriginalDst
}

func (al *activeListener) GetBindToPort() bool {
	return al.bindToPort
}

func (al *activeListener) AddAcceptFilter(f api.ListenerFilter) {
	al.filters = append(al.filters, f)
}

func (al *activeListener) Start() error {
	return al.listener.Listen()
}

func (al *activeListener) OnAccept(conn api.Connection) {
	log.Debug("[Listener: %s]", al.pb.GetName())

	lcb := &listenerCallbacks{conn}

	if !al.onListenerFilter(lcb) {
		conn.Close()
		return
	}

	if al.GetUseOriginalDst() {
		ip, port := conn.Context().GetDestinationIP(), conn.Context().GetDestinationPort()
		rdl := al.getRedirectListener(ip, port)
		//  If there is no listener associated with the original destination address,
		//  the connection is handled by the listener that receives it
		if rdl != nil {
			log.Debug("redirect listener to: %s", rdl.Listener().Addr().String())
			cb := rdl.(api.ListenerCallback)
			cb.OnAccept(conn)
			return
		}
		log.Debug("get redirect listener fail: %v %d", ip, port)
	}

	ac := NewActiveConnection(conn)
	filters := al.matchFilters(conn.Context())
	if filters == nil {
		log.Error("Match filter chain fail")
		conn.Close()
		return
	}

	for _, f := range filters {
		factory, pb := filter.GetNetworkFactory(f.GetTypedConfig(), f.Name)
		if factory == nil {
			if filter.IsWellknowName(f.Name) {
				panic(fmt.Errorf("not found network factory:%s", f.Name))
			} else {
				log.Error("not support network filter: %s", f.Name)
				continue
			}
		}
		if err := factory.CreateFilterFactory(pb, al.context)(ac, ac); err != nil {
			log.Error("create network filter fail: %s", err)
			conn.Close()
			return
		}
	}

	ac.OnLoop()
}

func (al *activeListener) onListenerFilter(cb api.ListenerFilterCallbacks) bool {
	for _, f := range al.filters {
		if f.OnAccept(cb) == api.Stop {
			return false
		}
	}
	return true
}

func (al *activeListener) matchFilters(cs api.ConnectionContext) []*envoy_config_listener_v3.Filter {
	filterChain := al.filterChainManager.FindFilterChains(cs)
	if filterChain != nil {
		log.Debug("[FilterChain: %s]", filterChain.GetName())
		return filterChain.Filters
	}
	return nil
}

func (al *activeListener) getRedirectListener(ip net.IP, port uint32) api.ActiveListener {
	addr := &net.TCPAddr{IP: ip, Port: int(port)}
	return al.context.ListenerManager().FindListenerByAddress(addr)
}

type listenerCallbacks struct {
	conn api.Connection
}

func (cb *listenerCallbacks) Connection() api.Connection {
	return cb.conn
}
