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
	"strings"

	envoy_config_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/wereliang/govoy/pkg/api"
	"github.com/wereliang/govoy/pkg/log"
	"github.com/yl2chen/cidranger"
)

var (
	EmptyPort                = uint32(0)
	EmptyIP                  = net.ParseIP("0.0.0.0")
	EmptyIPNet               = "0.0.0.0/0"
	EmptyServerName          = ""
	EmptyTransportProtocol   = "raw_buffer"
	EmptyApplicationProtocol = ""
)

type MatchTable interface {
	Find(k interface{}) (interface{}, bool)
	Insert(k interface{}, t interface{})
	Loop(func(k, v, arg interface{}), interface{})
	Len() int
}

type MapTable struct {
	maps map[interface{}]interface{}
}

func newMapTable() MatchTable {
	return &MapTable{maps: make(map[interface{}]interface{})}
}

func (mt *MapTable) Find(k interface{}) (interface{}, bool) {
	v, b := mt.maps[k]
	return v, b
}

func (mt *MapTable) Insert(k interface{}, t interface{}) {
	mt.maps[k] = t
}

func (mt *MapTable) Loop(cb func(interface{}, interface{}, interface{}), arg interface{}) {
	for k, v := range mt.maps {
		cb(k, v, arg)
	}
}

func (mt *MapTable) Len() int {
	return len(mt.maps)
}

type CirdEntry struct {
	cidranger.RangerEntry
	ipnet *net.IPNet
	table MatchTable
}

func (e *CirdEntry) Network() net.IPNet {
	return *e.ipnet
}

type CidrRangerTable struct {
	cidr cidranger.Ranger
	maps map[string]MatchTable
}

func newCidrRangerTable() MatchTable {
	return &CidrRangerTable{
		cidr: cidranger.NewPCTrieRanger(),
		maps: make(map[string]MatchTable),
	}
}

func (c *CidrRangerTable) Find(k interface{}) (interface{}, bool) {
	x, ok := c.maps[k.(string)]
	return x, ok
}

func (c *CidrRangerTable) ContainingNetworks(ip net.IP) (MatchTable, error) {
	entrys, err := c.cidr.ContainingNetworks(ip)
	if err != nil {
		return nil, err
	}
	last := entrys[len(entrys)-1]
	return last.(*CirdEntry).table, nil
}

func (c *CidrRangerTable) Insert(k interface{}, t interface{}) {
	skey := k.(string)
	_, network, _ := net.ParseCIDR(skey)
	// c.cidr.Insert(cidranger.NewBasicRangerEntry(*network))
	c.cidr.Insert(&CirdEntry{ipnet: network, table: t.(MatchTable)})
	c.maps[skey] = t.(MatchTable)
}

func (c *CidrRangerTable) Loop(cb func(interface{}, interface{}, interface{}), arg interface{}) {
	for k, v := range c.maps {
		cb(k, v, arg)
	}
}

func (c *CidrRangerTable) Len() int {
	return len(c.maps)
}

/*
filter_chains is array, every item included chain_match and filters，the first way is math
	filter_chains:
	- filter_chain_match
	  filters
	- filter_chain_match
      filters

The following order applies:

1 Destination port.
2 Destination IP address.
3 Server name (e.g. SNI for TLS protocol),
4 Transport protocol.
5 Application protocols (e.g. ALPN for TLS protocol).
6 Directly connected source IP address (this will only be different from the source IP address when using a listener filter that overrides the source address, such as the Proxy Protocol listener filter).
7 Source type (e.g. any, local or external network).
8 Source IP address.
9 Source port.
*/
type FilterChainManagerImpl struct {
	destPortsMap       MatchTable
	defaultFilterChain *envoy_config_listener_v3.FilterChain
}

func newFilterChainManager(
	filterChains []*envoy_config_listener_v3.FilterChain,
	defaultChain *envoy_config_listener_v3.FilterChain) api.FilterChainManager {

	filterChainManager := &FilterChainManagerImpl{
		destPortsMap: newMapTable(),
	}
	filterChainManager.AddFilterChains(filterChains, defaultChain)
	return filterChainManager
}

func (fm *FilterChainManagerImpl) AddFilterChains(
	filterChains []*envoy_config_listener_v3.FilterChain,
	defaultFilterChain *envoy_config_listener_v3.FilterChain) {

	for _, filterChain := range filterChains {
		if filterChain.GetFilterChainMatch() != nil {
			fm.addFilterChainForDestinationPorts(filterChain, filterChain.GetFilterChainMatch())
		}
	}

	fm.defaultFilterChain = defaultFilterChain
	// if not config match and default filter chain
	if defaultFilterChain == nil && fm.destPortsMap.Len() == 0 && len(filterChains) > 0 {
		fm.defaultFilterChain = filterChains[0]
		if fm.defaultFilterChain.GetName() == "" {
			fm.defaultFilterChain.Name = "only"
		}
	}
}

func (fm *FilterChainManagerImpl) addMatchTable(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch,
	table MatchTable,
	val interface{},
	nextTabFunc func() MatchTable,
	nextAddFunc func(*envoy_config_listener_v3.FilterChain, *envoy_config_listener_v3.FilterChainMatch, MatchTable)) {

	nextTab, ok := table.Find(val)
	if !ok {
		nextTab = nextTabFunc()
		table.Insert(val, nextTab)
	}
	nextAddFunc(filterChain, match, nextTab.(MatchTable))
}

func (fm *FilterChainManagerImpl) addFilterChainForDestinationPorts(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch) {

	fm.addMatchTable(
		filterChain,
		match,
		fm.destPortsMap,
		match.DestinationPort.GetValue(),
		newCidrRangerTable,
		fm.addFilterChainForDestinationIPs)
}

func (fm *FilterChainManagerImpl) addFilterChainForDestinationIPs(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch, maps MatchTable) {

	fn := func(s string) {
		fm.addMatchTable(
			filterChain,
			match,
			maps,
			s,
			newMapTable,
			fm.addFilterChainForServerNames)
	}
	if match.GetPrefixRanges() == nil {
		fn(EmptyIPNet)
	} else {
		for _, r := range match.GetPrefixRanges() {
			fn(fmt.Sprintf("%s/%d", r.GetAddressPrefix(), r.GetPrefixLen().GetValue()))
		}
	}
}

func (fm *FilterChainManagerImpl) addFilterChainForServerNames(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch, maps MatchTable) {

	fn := func(s string) {
		fm.addMatchTable(
			filterChain,
			match,
			maps,
			s,
			newMapTable,
			fm.addFilterChainForTransportPortocol)
	}
	if match.GetServerNames() == nil {
		fn(EmptyServerName)
	} else {
		for _, sname := range match.GetServerNames() {
			// not support wildchar now
			if strings.Contains(sname, "*") {
				panic(fmt.Sprintf("not support wildchar server name:%s", sname))
			}
			fn(sname)
		}
	}
}

func (fm *FilterChainManagerImpl) addFilterChainForTransportPortocol(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch, maps MatchTable) {
	protocol := match.GetTransportProtocol()
	if protocol == "" {
		protocol = EmptyTransportProtocol
	}
	fm.addMatchTable(
		filterChain,
		match,
		maps,
		protocol,
		newMapTable,
		fm.addFilterChainForApplicationProtocols)
}

func (fm *FilterChainManagerImpl) addFilterChainForApplicationProtocols(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch, maps MatchTable) {

	fn := func(s string) {
		fm.addMatchTable(
			filterChain,
			match,
			maps,
			s,
			newCidrRangerTable,
			fm.addFilterChainForDirectSourceIPs)
	}
	if match.GetApplicationProtocols() == nil {
		fn(EmptyApplicationProtocol)
	} else {
		for _, ap := range match.GetApplicationProtocols() {
			fn(ap)
		}
	}
}

func (fm *FilterChainManagerImpl) addFilterChainForDirectSourceIPs(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch, maps MatchTable) {

	fn := func(s string) {
		fm.addMatchTable(
			filterChain,
			match,
			maps,
			s,
			newMapTable,
			fm.addFilterChainForSourceType)
	}
	if match.GetDirectSourcePrefixRanges() == nil {
		fn(EmptyIPNet)
	} else {
		for _, r := range match.GetDirectSourcePrefixRanges() {
			fn(fmt.Sprintf("%s/%d", r.GetAddressPrefix(), r.GetPrefixLen().GetValue()))
		}
	}
}

func (fm *FilterChainManagerImpl) addFilterChainForSourceType(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch, maps MatchTable) {

	fm.addMatchTable(
		filterChain,
		match,
		maps,
		match.GetSourceType(),
		newCidrRangerTable,
		fm.addFilterChainForSourceIPs)
}

func (fm *FilterChainManagerImpl) addFilterChainForSourceIPs(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch, maps MatchTable) {

	fn := func(s string) {
		fm.addMatchTable(
			filterChain,
			match,
			maps,
			s,
			newMapTable,
			fm.addFilterChainForSourcePorts)
	}
	if match.GetSourcePrefixRanges() == nil {
		fn(EmptyIPNet)
	} else {
		for _, r := range match.GetSourcePrefixRanges() {
			fn(fmt.Sprintf("%s/%d", r.GetAddressPrefix(), r.GetPrefixLen().GetValue()))
		}
	}
}

func (fm *FilterChainManagerImpl) addFilterChainForSourcePorts(
	filterChain *envoy_config_listener_v3.FilterChain,
	match *envoy_config_listener_v3.FilterChainMatch, maps MatchTable) {

	if match.GetSourcePorts() == nil {
		maps.Insert(EmptyPort, filterChain)
	} else {
		for _, port := range match.GetSourcePorts() {
			maps.Insert(port, filterChain)
		}
	}
}

func (fm *FilterChainManagerImpl) FindFilterChains(ctx api.ConnectionContext) *envoy_config_listener_v3.FilterChain {
	if fm.destPortsMap.Len() != 0 {
		filterChain := fm.findFilterChainsForDestinationPorts(ctx, fm.destPortsMap)
		if filterChain != nil {
			return filterChain
		}
	}
	return fm.defaultFilterChain
}

func (fm *FilterChainManagerImpl) findFilterChainsForDestinationPorts(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	t, b := fm.destPortsMap.Find(ctx.GetDestinationPort())
	if b {
		return fm.findFilterChainsForDestinationIPs(ctx, t.(MatchTable))
	}
	t, b = fm.destPortsMap.Find(EmptyPort)
	if b {
		return fm.findFilterChainsForDestinationIPs(ctx, t.(MatchTable))
	}
	log.Trace("not found destport : %d", ctx.GetDestinationPort())
	return nil
}

func (fm *FilterChainManagerImpl) findFilterChainsForDestinationIPs(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	destIP := ctx.GetDestinationIP()
	if destIP == nil {
		destIP = EmptyIP
	}
	table, err := maps.(*CidrRangerTable).ContainingNetworks(destIP)
	if err != nil {
		log.Trace("not found destination ip : %s %s", destIP, err)
		return nil
	}
	return fm.findFilterChainsForServerNames(ctx, table)
}

func (fm *FilterChainManagerImpl) findFilterChainsForServerNames(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	t, b := maps.Find(ctx.GetServerName())
	if !b {
		log.Trace("not found server name: %s", ctx.GetServerName())
		return nil
	}
	return fm.findFilterChainForTransportPortocol(ctx, t.(MatchTable))
}

func (fm *FilterChainManagerImpl) findFilterChainForTransportPortocol(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	protocol := ctx.GetTransportProtocol()
	if protocol == "" {
		protocol = EmptyTransportProtocol
	}
	t, b := maps.Find(protocol)
	if !b {
		log.Trace("not found transport protocol: %s", protocol)
		return nil
	}
	return fm.findFilterChainForApplicationProtocols(ctx, t.(MatchTable))
}

func (fm *FilterChainManagerImpl) findFilterChainForApplicationProtocols(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	t, b := maps.Find(ctx.GetApplicationProtocol())
	if !b {
		t, b = maps.Find(EmptyApplicationProtocol)
		if !b {
			log.Trace("not found application protocol: %s and empty", ctx.GetApplicationProtocol())
			return nil
		}
	}

	return fm.findFilterChainForDirectSourceIPs(ctx, t.(MatchTable))
}

func (fm *FilterChainManagerImpl) findFilterChainForDirectSourceIPs(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	directSourceIP := ctx.GetDirectSourceIP()
	if directSourceIP == nil {
		directSourceIP = EmptyIP
	}
	t, err := maps.(*CidrRangerTable).ContainingNetworks(directSourceIP)
	if err != nil {
		log.Trace("not found directSourceIP : %s %s", directSourceIP, err)
		return nil
	}

	return fm.findFilterChainForSourceType(ctx, t)
}

func (fm *FilterChainManagerImpl) findFilterChainForSourceType(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	t, b := maps.Find(envoy_config_listener_v3.FilterChainMatch_ConnectionSourceType(ctx.GetSourceType()))
	if !b {
		log.Trace("not found SourceType: %v", ctx.GetSourceType())
		return nil
	}
	return fm.findFilterChainForSourceIPs(ctx, t.(MatchTable))
}

func (fm *FilterChainManagerImpl) findFilterChainForSourceIPs(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	sourceIP := ctx.GetSourceIP()
	if sourceIP == nil {
		sourceIP = EmptyIP
	}
	t, err := maps.(*CidrRangerTable).ContainingNetworks(sourceIP)
	if err != nil {
		log.Trace("not found sourceIP : %s %s", sourceIP, err)
		return nil
	}
	return fm.findFilterChainForSourcePorts(ctx, t)
}

func (fm *FilterChainManagerImpl) findFilterChainForSourcePorts(
	ctx api.ConnectionContext, maps MatchTable) *envoy_config_listener_v3.FilterChain {

	filterChain, b := maps.Find(ctx.GetSourcePort())
	if !b {
		filterChain, b = maps.Find(EmptyPort)
	}
	if b {
		return filterChain.(*envoy_config_listener_v3.FilterChain)
	}
	log.Trace("not found SourcePort: %v", ctx.GetSourcePort())
	return nil
}
